package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/deweppro/go-errors"
	"github.com/deweppro/go-utils/routine"
)

type (
	sh struct {
		env   []string
		dir   string
		shell string
		mux   sync.RWMutex
		w     io.Writer
		ch    chan []byte
	}

	Shell interface {
		Close()
		SetEnv(key, value string)
		SetDir(dir string)
		SetShell(shell string)
		SetWriter(w io.Writer)
		CallPackageContext(ctx context.Context, commands ...string) error
		CallParallelContext(ctx context.Context, commands ...string) error
		CallContext(ctx context.Context, command string) error
		Call(ctx context.Context, command string) ([]byte, error)
	}
)

func New() Shell {
	v := &sh{
		env:   make([]string, 0),
		dir:   os.TempDir(),
		shell: "/bin/sh",
		w:     &NullWriter{},
		ch:    make(chan []byte, 128),
	}
	go v.Pipe()
	return v
}

func (v *sh) SetEnv(key, value string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.env = append(v.env, key+"="+value)
}

func (v *sh) SetDir(dir string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.dir = dir
}

func (v *sh) SetShell(shell string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.shell = shell
}

func (v *sh) SetWriter(w io.Writer) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.w = w
}

func (v *sh) Close() {
	v.SetWriter(&NullWriter{})
	close(v.ch)
}

func (v *sh) Pipe() {
	for {
		b, ok := <-v.ch
		if !ok {
			return
		}
		v.mux.RLock()
		var bb []byte
		copy(bb, b)
		v.w.Write(bb) //nolint:errcheck
		v.mux.RUnlock()
	}
}

func (v *sh) Write(b []byte) (n int, err error) {
	l := len(b)
	select {
	case v.ch <- b:
	default:
	}
	return l, nil
}

func (v *sh) CallPackageContext(ctx context.Context, commands ...string) error {
	for i, command := range commands {
		if err := v.CallContext(ctx, command); err != nil {
			return errors.WrapMessage(err, "call command #%d [%s]", i, command)
		}
	}
	return nil
}

func (v *sh) CallParallelContext(ctx context.Context, commands ...string) error {
	var (
		err error
		mux sync.Mutex
	)
	calls := make([]func(), 0, len(commands))
	for i, c := range commands {
		c := c
		calls = append(calls, func() {
			if e := v.CallContext(ctx, c); e != nil {
				mux.Lock()
				err = errors.Wrap(err, errors.WrapMessage(err, "call command #%d [%s]", i, c))
				mux.Unlock()
			}
		})
	}
	routine.Parallel(calls...)
	if err != nil {
		return err
	}
	return nil
}

func (v *sh) CallContext(ctx context.Context, c string) error {
	v.mux.RLock()
	if _, err := fmt.Fprintf(v, c+"\n"); err != nil {
		v.mux.RUnlock()
		return err
	}

	cmd := exec.CommandContext(ctx, v.shell, "-xec", c, " <&-")
	cmd.Env = append(os.Environ(), v.env...)
	cmd.Dir = v.dir
	cmd.Stdout = v
	cmd.Stderr = v
	v.mux.RUnlock()

	return cmd.Run()
}

func (v *sh) Call(ctx context.Context, c string) ([]byte, error) {
	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", c, " <&-")
	cmd.Env = append(os.Environ(), v.env...)
	cmd.Dir = v.dir
	v.mux.RUnlock()

	return cmd.CombinedOutput()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type NullWriter struct {
}

func (v *NullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
