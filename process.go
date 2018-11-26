package gowatch

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/pkg/errors"
)

// Process controls execution of several goroutines
type Process struct {
	mu     sync.Mutex
	errors chan error
	stop   chan struct{}

	// watchedCount current amount of watchedCount goroutines
	watchedCount int64

	// StopSignals contains list of system signals which will stop application
	// if left empty SIGINT and SIGTERM signals will be considered stopping
	StopSignals []os.Signal
	lastErr     error
}

// Run starts new process as a goroutine
func (p *Process) Run(proc func()) {
	go p.runE(func() error {
		proc()

		return nil
	})
}

// RunE starts new process as a goroutine and watch for errors
func (p *Process) RunE(proc func() error) {
	go p.runE(proc)
}

// RunStop starts new process as a goroutine and watch for errors
// Passes stop channel to function, which will be closed
func (p *Process) RunStop(proc func(stop <-chan struct{})) {
	p.initStopCh()
	atomic.AddInt64(&p.watchedCount, 1)
	go p.runE(func() error {
		defer atomic.AddInt64(&p.watchedCount, -1)
		proc(p.stop)

		return nil
	})
}

// RunStopE starts new process as a goroutine
// Passes stop channel to function, which will be closed
func (p *Process) RunStopE(proc func(stop <-chan struct{}) error) {
	p.initStopCh()
	atomic.AddInt64(&p.watchedCount, 1)
	go p.runE(func() error {
		defer atomic.AddInt64(&p.watchedCount, -1)
		return proc(p.stop)
	})
}

// Watch start watching for started goroutines
// returns when one of goroutines stopped, if goroutine exited with error
// error is returned
func (p *Process) Watch() error {
	p.initErrCh()

	p.initStopCh()
	signals := make(chan os.Signal)
	if len(p.StopSignals) == 0 {
		p.StopSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	signal.Notify(signals, p.StopSignals...)

	go func() {
		<-signals
		p.doStop()
	}()

	return p.waitForErr()
}

func (p *Process) runE(proc func() error) {
	defer func() {
		if err := recover(); err != nil {
			p.watchErr(errors.Errorf("panic: %v", err))
		}
	}()

	p.watchErr(proc())
}

func (p *Process) watchErr(err error) {
	p.initErrCh()

	p.errors <- err
}

func (p *Process) initStopCh() {
	p.mu.Lock()
	if p.stop == nil {
		p.stop = make(chan struct{})
	}
	p.mu.Unlock()
}

func (p *Process) initErrCh() {
	p.mu.Lock()
	if p.errors == nil {
		p.errors = make(chan error, 1)
	}
	p.mu.Unlock()
}

func (p *Process) waitForErr() error {
	p.initErrCh()
	for {
		if err := <-p.errors; err != nil {
			p.doStop()
			if err != nil {
				p.lastErr = err
			}
		}

		if wc := atomic.LoadInt64(&p.watchedCount); wc == 0 {
			return p.lastErr
		}
	}
}

// Stop start stop sequence
func (p *Process) Stop() {
	p.doStop()
}

func (p *Process) doStop() {
	select {
	case <-p.stop: // no one writes to this channel
	default:
		close(p.stop)
	}
}