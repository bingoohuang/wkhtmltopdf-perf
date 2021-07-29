package wkhtml

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bingoohuang/wkp/pkg/util"
)

func (p *ToX) ToPdf(htmlURL, extraArgs string, saveFile bool) (pdf []byte, err error) {
	var out string
	if out, err = util.TempFile(".pdf"); err != nil {
		return
	}
	if !saveFile {
		defer os.Remove(out)
	}

	cmd := wkhtmltopdf + " " + extraArgs + p.CacheDirArg() + " --quiet " + strconv.Quote(htmlURL) + " " + out
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	_, err = options.Exec(nil, "sh", "-c", cmd)
	if err == nil {
		if saveFile {
			return []byte(out), nil
		}
		return os.ReadFile(out)
	}

	return nil, err
}
