package wkhtml

import (
	"log"
	"time"
)

func (p *ToX) ToPdfV1(url, extraArgs string) (pdf []byte, err error) {
	data, err := GetContent(url)
	if err != nil {
		return nil, err
	}

	log.Printf("content read (%d): %s", len(data), url)

	cmd := wkhtmltopdf + " " + extraArgs + " --quiet " + " - - | cat"
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(data, "sh", "-c", cmd)
}
