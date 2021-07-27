package wkhtml

import (
	"bytes"
	"github.com/bingoohuang/wkp/pkg/util"
	"io"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func (p *ToX) ToPdfV1p(url, _ string) (pdf []byte, err error) {
	data, err := util.GetContent(url)
	if err != nil {
		return nil, err
	}

	v1pOnce.Do(func() { v1pPool = NewV1pPool() })

	item := v1pPool.borrow()
	return item.Exec(data)
}

var v1pPool *V1pPool
var v1pOnce sync.Once

type V1pPool struct {
	chn chan *V1pItem
}

func NewV1pPool() *V1pPool {
	options := ExecOptions{Timeout: 10 * time.Second}
	p := &V1pPool{}
	p.chn = make(chan *V1pItem, runtime.NumCPU()*2)
	go func() {
		cmd := wkhtmltopdf + " --cache-dir /tmp/cache-wk/ " + extra + " --quiet - -|cat"
		for {
			wk, err := options.NewV1pItem("sh", "-c", cmd)
			if err != nil {
				time.Sleep(10 * time.Second)
				continue
			}
			p.chn <- wk
		}
	}()

	return p
}

func (p *V1pPool) borrow() *V1pItem { return <-p.chn }

const extra = `--dpi 96 --print-media-type --page-size A4 --orientation Landscape --margin-top 10mm --margin-bottom 10mm` +
	` --margin-left 10mm --margin-right 10mm --footer-center '[page]/[topage]'`

type V1pItem struct {
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout, stderr io.ReadCloser
	timeout        time.Duration
}

func (o ExecOptions) NewV1pItem(name string, args ...string) (p *V1pItem, err error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}

	p = &V1pItem{timeout: o.Timeout}
	p.cmd = cmd
	if p.stdin, err = cmd.StdinPipe(); err != nil {
		return
	}
	if p.stdout, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if p.stderr, err = cmd.StderrPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	return
}

func (p *V1pItem) Exec(data []byte) (result []byte, err error) {
	if _, err = io.Copy(p.stdin, bytes.NewBuffer(data)); err != nil {
		return
	}

	p.stdin.Close()

	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	go io.Copy(outBuf, p.stdout)
	go io.Copy(errBuf, p.stderr)

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- p.cmd.Wait()
	}()

	select {
	case err = <-ch:
	case <-time.After(p.timeout):
		p.cmd.Process.Kill()
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
