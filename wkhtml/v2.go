package wkhtml

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type V2Pool struct {
	ch1, chn chan *V2Item
}

func NewV2Pool() *V2Pool {
	options := ExecOptions{Timeout: 10 * time.Second}
	p := &V2Pool{}
	p.ch1 = make(chan *V2Item)
	p.chn = make(chan *V2Item, runtime.NumCPU()*2)
	go func() {
		for {
			wk, err := options.NewV2Item(wkhtmltopdf, "--read-args-from-stdin")
			if err != nil {
				time.Sleep(10 * time.Second)
				continue
			}
			p.ch1 <- wk
		}
	}()

	return p
}

func (p *V2Pool) borrow() *V2Item {
	select {
	case wk := <-p.chn:
		return wk
	default:
		return <-p.ch1
	}
}

func (p *V2Pool) back(wk *V2Item) {
	select {
	case p.chn <- wk:
		return
	default:
		if err := wk.Kill(); err != nil {
			log.Printf("failed to kill, error: %v", err)
		}
		return
	}
}

var v2Pool = NewV2Pool()

func (p *ToX) ToPdfV2(htmlURL, extraArgs string) (pdf []byte, err error) {
	var out string
	if out, err = CreateTempFile(); err != nil {
		return
	}
	defer os.Remove(out)

	in := strconv.Quote(htmlURL) + " " + out + "\n"
	if extraArgs != "" {
		in = extraArgs + " " + in
	}

	wk := v2Pool.borrow()
	result, err := wk.Send(in, "Done", "Error:")
	log.Printf("wk result: %s", result)
	if err == ErrTimeout {
		if err := wk.Kill(); err != nil {
			log.Printf("failed to kill, error: %v", err)
		}
	} else {
		v2Pool.back(wk)
	}

	if err == nil {
		return os.ReadFile(out)
	}

	return nil, err
}

type V2Item struct {
	In, Out    chan string
	cmd        *exec.Cmd
	Timeout    time.Duration
	StdoutPipe io.ReadCloser
}

func (i *V2Item) Send(input string, okTerm, errTerm string) (string, error) {
	ClearChan(i.Out)
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

func (i *V2Item) Kill() interface{} {
	i.StdoutPipe.Close()
	return i.cmd.Process.Kill()
}

func (o ExecOptions) NewV2Item(name string, args ...string) (inOut *V2Item, err error) {
	inOut = &V2Item{In: make(chan string), Out: make(chan string), Timeout: o.Timeout}
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
