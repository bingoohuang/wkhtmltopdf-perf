package wkhtml

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/bingoohuang/wkp/pkg/util"
)

type ToX struct {
	MaxPoolSize int
	CacheDir    bool

	//  以下三项，适用于 WkVersion 为 2 或 2p 的情况，用于判断转换是否成功
	OkItems []string
	Timeout time.Duration
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
	options := ExecOptions{Timeout: p.Timeout}
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

var (
	ErrTimeout = errors.New("execute timeout")
	ErrExecute = errors.New("execute error")
)

func (o ExecOptions) Exec(data []byte, name string, args ...string) (result []byte, err error) {
	item, err := o.NewV1pItem(name, args...)
	if err != nil {
		return nil, err
	}
	return item.Exec(data)
}
