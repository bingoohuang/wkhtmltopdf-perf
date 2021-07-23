package wkhtml

import (
	"time"
)

type ToX struct {
}

func (p *ToX) ToPDF(html []byte) (pdf []byte, err error) {
	args := []string{"--quiet", "-", "-"}
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(html, "wkhtmltopdf", args...)
}
