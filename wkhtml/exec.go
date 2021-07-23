package wkhtml

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"syscall"
	"time"
)

type ExecOptions struct {
	Timeout time.Duration
}

var ErrTimeout = errors.New("execute timeout")

func (o ExecOptions) Exec(data []byte, name string, args ...string) (result []byte, err error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	if err = cmd.Start(); err != nil {
		return
	}

	if _, err = io.Copy(stdin, bytes.NewBuffer(data)); err != nil {
		return
	}

	stdin.Close()

	go io.Copy(outBuf, stdout)
	go io.Copy(errBuf, stderr)

	ch := make(chan error)
	go func(cmd *exec.Cmd) {
		defer close(ch)
		ch <- cmd.Wait()
	}(cmd)

	select {
	case err = <-ch:
	case <-time.After(o.Timeout):
		cmd.Process.Kill()
		err = ErrTimeout
		return
	}

	if err != nil {
		return nil, err
	}

	return outBuf.Bytes(), nil
}
