package wkhtml

import (
	"log"
	"time"

	"github.com/bingoohuang/wkp/pkg/util"
)

func (p *ToX) ToPdfV1(url, extraArgs string, saveFile bool) (pdf []byte, err error) {
	data, err := util.GetContent(url)
	if err != nil {
		return nil, err
	}

	log.Printf("content read (%d): %s", len(data), url)

	cmd := wkhtmltopdf + " " + extraArgs + p.CacheDirArg()
	var out string

	if saveFile {
		if out, err = util.TempFile(".pdf"); err != nil {
			return
		}
		cmd += " --quiet - " + out
	} else {
		cmd += " --quiet - - | cat"
	}
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	stdout, err := options.Exec(data, "sh", "-c", cmd)
	if err != nil {
		return nil, err
	}

	if saveFile {
		return []byte(out), nil
	}

	return stdout, nil
}
