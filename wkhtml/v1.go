package wkhtml

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func (p *ToX) ToPdfV1(url, extraArgs string) (pdf []byte, err error) {
	data, err := GetContent(url)
	if err != nil {
		return nil, err
	}

	cmd := wkhtmltopdf + " " + extraArgs + " --quiet " + " - - | cat"
	log.Printf("cmd: %s", cmd)
	options := ExecOptions{Timeout: 10 * time.Second}
	return options.Exec(data, "sh", "-c", cmd)
}

func GetContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
