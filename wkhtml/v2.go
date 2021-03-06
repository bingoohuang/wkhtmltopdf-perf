package wkhtml

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bingoohuang/gg/pkg/ss"

	"github.com/bingoohuang/wkp/pkg/util"
	"go.uber.org/multierr"
)

type V2Pool struct {
	ch       chan *V2Item
	wait     chan bool
	num, max int32
}

func NewV2Pool(tox *ToX) *V2Pool {
	max := tox.MaxPoolSize
	options := ExecOptions{Timeout: tox.Timeout}
	p := &V2Pool{max: int32(max), ch: make(chan *V2Item, max), wait: make(chan bool)}
	go func() {
		for {
			<-p.wait
			if atomic.LoadInt32(&p.num) >= p.max {
				continue // 生产已经达到上限
			}

			wk, err := options.NewV2Item(tox, wkhtmltopdf, "--read-args-from-stdin")
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			atomic.AddInt32(&p.num, 1)
			p.ch <- wk
		}
	}()

	return p
}

func (p *V2Pool) borrow() *V2Item {
	// 尝试获取
	select {
	case v := <-p.ch:
		return v
	default:
	}
	// 通知生产
	select {
	case p.wait <- true:
	default:
	}
	// 取走
	return <-p.ch
}

func (p *V2Pool) back(wk *V2Item) {
	if wk != nil {
		p.ch <- wk
		return
	}

	// report missing
	atomic.AddInt32(&p.num, -1)
}

var (
	v2Pool *V2Pool
	v2Once sync.Once
)

func (p *ToX) ToPdfV2(htmlURL, extraArgs string, saveFile bool) (pdf []byte, err error) {
	v2Once.Do(func() { v2Pool = NewV2Pool(p) })
	var out string
	if out, err = util.TempFile(".pdf"); err != nil {
		return
	}
	if !saveFile {
		defer os.Remove(out)
	}

	err = p.SendArgs(htmlURL, extraArgs, out)
	if err != nil {
		return nil, err
	}
	if !saveFile {
		return os.ReadFile(out)
	}

	return []byte(out), nil
}

type V2Item struct {
	In, Out    chan string
	cmd        *exec.Cmd
	Timeout    time.Duration
	StdoutPipe io.ReadCloser
	Tox        *ToX
}

func (i *V2Item) Send(input string) (string, error) {
	util.ClearChan(i.Out)
	log.Printf("Send args: %s", input)
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
			if ss.Contains(line, i.Tox.OkItems...) {
				return out, nil
			}
		case <-time.After(i.Timeout):
			return out, ErrTimeout
		}
	}
}

func (i *V2Item) Kill(reason string) error {
	log.Printf("start to kill %d by %s", i.cmd.Process.Pid, reason)
	err1 := i.StdoutPipe.Close()
	err2 := i.cmd.Process.Kill()
	_, err3 := i.cmd.Process.Wait()
	return multierr.Combine(err1, err2, err3)
}

func (o ExecOptions) NewV2Item(tox *ToX, name string, args ...string) (inOut *V2Item, err error) {
	inOut = &V2Item{In: make(chan string), Out: make(chan string), Timeout: o.Timeout, Tox: tox}
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
