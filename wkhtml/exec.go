package wkhtml

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type ExecOptions struct {
	Timeout time.Duration
}

var ErrTimeout = errors.New("execute timeout")
var ErrExecute = errors.New("execute error")

type InOut struct {
	In, Out    chan string
	cmd        *exec.Cmd
	Timeout    time.Duration
	StdoutPipe io.ReadCloser
}

func (i *InOut) Send(input string, okTerm, errTerm string) (string, error) {
	clearOut(i.Out)
	i.In <- input

	out := ""
	for {
		select {
		case line := <-i.Out:
			if p := strings.LastIndexAny(line, "\r\n"); p > 0 {
				line = line[p+1:]
			}
			line = strings.TrimSpace(line)
			out += line
			if strings.Contains(line, okTerm) {
				return out, nil
			}
			if strings.Contains(line, errTerm) {
				return out, ErrExecute
			}
		case <-time.After(i.Timeout):
			return out, ErrTimeout
		}
	}
}

func (i *InOut) Kill() interface{} {
	i.StdoutPipe.Close()
	return i.cmd.Process.Kill()
}

func clearOut(out chan string) {
	for {
		select {
		case <-out:
		default:
			return
		}
	}
}

func (o ExecOptions) NewPrepare(name string, args ...string) (inOut *InOut, err error) {
	inOut = &InOut{In: make(chan string), Out: make(chan string), Timeout: o.Timeout}
	cmd := exec.Command(name, args...)
	inOut.cmd = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	// Make a new channel which will be used to ensure we get all output
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer log.Printf("exiting StdinPipe loop")
		for {
			select {
			case input, ok := <-inOut.In:
				if !ok {
					return
				}
				stdin.Write([]byte(input))
			case <-ctx.Done():
				return
			}
		}
	}()

	// Get a pipe to read from standard out
	inOut.StdoutPipe, _ = cmd.StdoutPipe()
	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {
		defer log.Printf("exiting StdoutPipe scanning loop")
		defer cancelFunc()

		// Create a scanner which scans r in a line-by-line fashion
		// Read line by line and process it
		for c := bufio.NewScanner(inOut.StdoutPipe); c.Scan(); {
			line := c.Text()
			inOut.Out <- line
		}
	}()

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	return inOut, nil
}

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
		log.Printf("Error: %s", err.Error())
		log.Printf("Stderr: %s", errBuf.String())
		return nil, err
	}

	return outBuf.Bytes(), nil
}
