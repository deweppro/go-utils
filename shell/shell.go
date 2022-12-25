package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	errors "github.com/deweppro/go-errors"
	"github.com/deweppro/go-utils/routine"
)

type (
	Shell struct {
		env   []string
		dir   string
		shell string
		mux   sync.RWMutex
		w     Writer
	}

	Writer func(s string)
)

func New(w Writer) *Shell {
	return &Shell{
		env:   make([]string, 0),
		dir:   os.TempDir(),
		shell: "/bin/sh",
		w:     w,
	}
}

func (v *Shell) SetEnv(key, value string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.env = append(v.env, key+"="+value)
}

func (v *Shell) SetDir(dir string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.dir = dir
}

func (v *Shell) SetShell(shell string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.shell = shell
}

func (v *Shell) CallPackageContext(ctx context.Context, commands ...string) error {
	for i, command := range commands {
		if err := v.CallContext(ctx, command); err != nil {
			return errors.WrapMessage(err, "call command #%d [%s]", i, command)
		}
	}
	return nil
}

func (v *Shell) CallParallelContext(ctx context.Context, commands ...string) error {
	var err error
	calls := make([]func(), 0, len(commands))
	for i, command := range commands {
		calls = append(calls, func() {
			if e := v.CallContext(ctx, command); e != nil {
				err = errors.Wrap(err, errors.WrapMessage(err, "call command #%d [%s]", i, command))
			}
		})
	}
	routine.Parallel(calls...)
	if err != nil {
		return err
	}
	return nil
}

func (v *Shell) CallContext(ctx context.Context, command string) error {
	v.w(command)

	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", fmt.Sprintln(command, " <&-"))
	cmd.Env = append(v.env, os.Environ()...)
	cmd.Dir = v.dir
	v.mux.RUnlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WrapMessage(err, "stdout init")
	}
	defer stdout.Close() //nolint: errcheck

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WrapMessage(err, "stderr init")
	}
	defer stderr.Close() //nolint: errcheck

	if err = cmd.Start(); err != nil {
		return errors.WrapMessage(err, "start command")
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			v.w(scanner.Text())
			select {
			case <-ctx.Done():
				break
			default:
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			v.w(scanner.Text())
			select {
			case <-ctx.Done():
				break
			default:
			}
		}
	}()

	return cmd.Wait()
}
