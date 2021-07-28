package wkhtml

import (
	"errors"
	"github.com/bingoohuang/wkp/pkg/util"
	"log"
	"strconv"
	"time"
)

type ToX struct {
	MaxPoolSize int
	CacheDir    bool
}

const wkhtmltopdf = "wkhtmltopdf"

func (p *ToX) ToPdfV0(url, extraArgs string, saveFile bool) (pdf []byte, err error) {
	cmd := wkhtmltopdf + " " + extraArgs + p.CacheDirArg() + " --quiet " + strconv.Quote(url)
	var out string

	if saveFile {
		if out, err = util.TempFile(".pdf"); err != nil {
			return
		}
		cmd += " " + out
	} else {
		cmd += " - | cat"
	}
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	output, err := options.Exec(nil, "sh", "-c", cmd)
	if err != nil {
		return nil, err
	}

	if saveFile {
		return []byte(out), nil
	}

	return output, nil
}

func (p *ToX) CacheDirArg() string {
	if !p.CacheDir {
		return ""
	}

	return " --cache-dir /tmp/cache-wk/ "
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
