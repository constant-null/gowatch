package gowatch

import (
	"reflect"
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
