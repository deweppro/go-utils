package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	errors "github.com/deweppro/go-errors"
	"github.com/deweppro/go-utils/routine"
)

const eof = byte('\n')

type (
	Shell struct {
		env   []string
		dir   string
		shell string
		mux   sync.RWMutex
		w     io.Writer
	}
)

func New() *Shell {
	return &Shell{
		env:   make([]string, 0),
		dir:   os.TempDir(),
		shell: "/bin/sh",
		w:     &NullWriter{},
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

func (v *Shell) SetWriter(w io.Writer) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.w = w
}

func (v *Shell) writeBytes(b []byte) {
	v.w.Write(append(b, eof)) //nolint: errcheck
}

func (v *Shell) writeString(b string) {
	v.w.Write(append([]byte(b), eof)) //nolint: errcheck
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
		command := command
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
	v.writeString(command)

	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", fmt.Sprintln(command, " <&-"))
	cmd.Env = append(os.Environ(), v.env...)
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
			v.writeBytes(scanner.Bytes())
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
			v.writeBytes(scanner.Bytes())
			select {
			case <-ctx.Done():
				break
			default:
			}
		}
	}()

	return cmd.Wait()
}

func (v *Shell) Call(ctx context.Context, command string) ([]byte, error) {
	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", fmt.Sprintln(command, " <&-"))
	cmd.Env = append(os.Environ(), v.env...)
	cmd.Dir = v.dir
	v.mux.RUnlock()

	return cmd.CombinedOutput()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type NullWriter struct {
}

func (v *NullWriter) Write(_ []byte) (int, error) {
	return 0, nil
}
