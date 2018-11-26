package gowatch

import (
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
)

var errTestTimeout = errors.New("test timeout")

func TestProgram_RunE(t *testing.T) {
	var p Process
	var errExpected = errors.New("expected error")

	p.RunE(func() error {
		time.Sleep(3 * time.Second)
		return errTestTimeout
	})

	p.RunE(func() error {
		return errExpected
	})

	err := p.Watch()
	if err != errExpected {
		t.Errorf("test failed: %s", err)
	}
}

func TestProgram_RunPanic(t *testing.T) {
	var p Process
	var errExpected = errors.New("expected error")

	p.RunE(func() error {
		time.Sleep(3 * time.Second)
		return errTestTimeout
	})

	p.RunE(func() error {
		panic(errExpected)
	})

	err := p.Watch()
	if !reflect.DeepEqual(err.Error(), errors.Errorf("panic: %v", errExpected).Error()) {
		t.Errorf("test failed: %s", err)
	}
}

func TestProgram_RunNilError(t *testing.T) {
	var p Process
	var errExpected error

	p.RunE(func() error {
		time.Sleep(3 * time.Second)
		return errTestTimeout
	})

	p.RunE(func() error {
		return nil
	})

	err := p.Watch()
	if err != errExpected {
		t.Errorf("test failed: %s", err)
	}
}

func TestProgram_RunStopE(t *testing.T) {
	var errExpected error = errors.New("application stopped")
	p := Process{StopSignals: []os.Signal{syscall.SIGTERM}}

	p.RunStopE(func(stop <-chan struct{}) error {
		select {
		case <-stop:
			return errExpected
		case <-time.After(3 * time.Second):
			return errTestTimeout
		}
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := p.Watch()
	if err != errExpected {
		t.Errorf("test failed: %s", err)
	}
}
