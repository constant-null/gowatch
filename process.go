package gowatch

import (
	"sync"

	"github.com/pkg/errors"
)

// Program controls execution of several goroutines
type Program struct {
	mu     sync.Mutex
	errors chan error
}

// Run starts new process as a goroutine
func (p *Program) Run(proc func()) {
	p.runE(func() error {
		proc()

		return nil
	})
}

// RunE starts new process as a goroutine and watch for errors
func (p *Program) RunE(proc func() error) {
	go p.runE(proc)
}

// Watch start watching for started goroutines
// returns when one of goroutines stopped, if goroutine exited with error
// error is returned
func (p *Program) Watch() error {
	p.initErrCh()

	return p.waitForErr()
}

func (p *Program) runE(proc func() error) {
	defer func() {
		if err := recover(); err != nil {
			p.watchErr(errors.Errorf("panic: %v", err))
		}
	}()

	p.watchErr(proc())
}

func (p *Program) watchErr(err error) {
	p.initErrCh()

	p.errors <- err
}

func (p *Program) initErrCh() {
	p.mu.Lock()
	if p.errors == nil {
		p.errors = make(chan error, 1)
	}
	p.mu.Unlock()
}

func (p *Program) waitForErr() error {
	p.initErrCh()
	return <-p.errors
}
