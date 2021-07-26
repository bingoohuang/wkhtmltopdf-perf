package wkhtml

import (
	"log"
	"os"
	"strconv"
	"time"
)

func (p *ToX) ToPdfV0(htmlURL, extraArgs string) (pdf []byte, err error) {
	var out string
	if out, err = createTemp(); err != nil {
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
