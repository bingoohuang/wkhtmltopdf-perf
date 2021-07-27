package wkhtml

import (
	"github.com/bingoohuang/wkp/pkg/util"
	"log"
	"os"
	"strconv"
	"time"
)

func (p *ToX) ToPdf(htmlURL, extraArgs string) (pdf []byte, err error) {
	var out string
	if out, err = util.TempFile(".pdf"); err != nil {
		return
	}
	defer os.Remove(out)

	cmd := wkhtmltopdf + " " + extraArgs + " --quiet " + strconv.Quote(htmlURL) + " " + out
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	_, err = options.Exec(nil, "sh", "-c", cmd)
	if err == nil {
		return os.ReadFile(out)
	}

	return nil, err
}
