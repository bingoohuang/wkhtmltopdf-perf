package wkhtml

import (
	"errors"
	"log"
	"strconv"
	"time"
)

type ToX struct {
}

const wkhtmltopdf = "wkhtmltopdf"

func (p *ToX) ToPdfV0(url, extraArgs string) (pdf []byte, err error) {
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
	item, err := o.NewV1pItem(name, args...)
	if err != nil {
		return nil, err
	}
	return item.Exec(data)
}
