package wkhtml

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

type ToX struct {
}

const wkhtmltopdf = "wkhtmltopdf"

func (p *ToX) ToPdf(url, extraArgs string) (pdf []byte, err error) {
	cmd := wkhtmltopdf + " " + extraArgs + " --quiet " + strconv.Quote(url) + " - | cat"
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(nil, "sh", "-c", cmd)
}

type ExecOptions struct {
	Timeout time.Duration
}

var ErrTimeout = errors.New("execute timeout")
var ErrExecute = errors.New("execute error")

func (o ExecOptions) Exec(data []byte, name string, args ...string) (result []byte, err error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}

	var stdin io.WriteCloser
	var stdout, stderr io.ReadCloser

	if stdin, err = cmd.StdinPipe(); err != nil {
		return
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	if _, err = io.Copy(stdin, bytes.NewBuffer(data)); err != nil {
		return
	}

	stdin.Close()

	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
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
		log.Printf("Error: %s", err.Error())
		log.Printf("Stderr: %s", errBuf.String())
		return nil, err
	}

	return outBuf.Bytes(), nil
}
