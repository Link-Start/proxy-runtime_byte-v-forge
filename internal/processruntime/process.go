package processruntime

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Path   string
	Args   []string
	Dir    string
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
}

type Process struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	exited bool
	err    error
}

func Start(ctx context.Context, cfg Config) (*Process, error) {
	processCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(processCtx, cfg.Path, cfg.Args...)
	cmd.Dir = cfg.Dir
	if len(cfg.Env) > 0 {
		cmd.Env = cfg.Env
	}
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	process := &Process{cmd: cmd, cancel: cancel, done: make(chan struct{})}
	go func() {
		err := cmd.Wait()
		process.mu.Lock()
		process.exited = true
		process.err = err
		process.mu.Unlock()
		close(process.done)
	}()
	return process, nil
}

func (p *Process) Done() <-chan struct{} {
	if p == nil {
		done := make(chan struct{})
		close(done)
		return done
	}
	return p.done
}

func (p *Process) Exited() bool {
	if p == nil {
		return true
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.exited
}

func (p *Process) ExitError() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *Process) Signal(signal os.Signal) error {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return errors.New("process is not running")
	}
	if p.Exited() {
		return errors.New("process is not running")
	}
	return p.cmd.Process.Signal(signal)
}

func (p *Process) Stop(timeout time.Duration) error {
	if p == nil {
		return nil
	}
	if p.Exited() {
		p.cancel()
		return nil
	}
	_ = p.Signal(syscall.SIGTERM)
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-p.done:
		p.cancel()
		return nil
	case <-timer.C:
		p.cancel()
		if p.cmd != nil && p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
		}
		<-p.done
		return nil
	}
}
