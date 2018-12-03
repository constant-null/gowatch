package gowatch

import (
	"sync"

	"github.com/pkg/errors"
)

// Process controls execution of several goroutines
type Process struct {
	mu     sync.Mutex
	errors chan error
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

// Watch start watching for started goroutines
// returns when one of goroutines stopped, if goroutine exited with error
// error is returned
func (p *Process) Watch() error {
	p.initErrCh()

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

func (p *Process) initErrCh() {
	p.mu.Lock()
	if p.errors == nil {
		p.errors = make(chan error, 1)
	}
	p.mu.Unlock()
}

func (p *Process) waitForErr() error {
	p.initErrCh()
	return <-p.errors
}
