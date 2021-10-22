package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/wkp/pkg/uuid"
)

func OrDuration(a, b time.Duration) time.Duration {
	if a == 0 {
		return b
	}

	return a
}

func OrSlice(a, b []string) []string {
	if len(a) > 0 {
		return a
	}

	return b
}

func GetContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func ClearChan(out chan string) {
	for {
		select {
		case <-out:
		default:
			return
		}
	}
}

func TempDir() (string, error) {
	return ioutil.TempDir("", "")
}

func TempFile(ext string) (string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	out := filepath.Join(dir, uuid.New().String()+ext)
	return out, nil
}

func FileData(name string) string {
	data, _ := os.ReadFile(name)
	return string(data)
}

func FileExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func ParseUploadFile(r *http.Request) (fn string, data []byte, err error) {
	if err = r.ParseMultipartForm(16 /*16 MiB */ << 20); err != nil {
		if err == http.ErrNotMultipart {
			err = nil
		}
		return "", nil, err
	}

	if r.MultipartForm == nil {
		return "", nil, nil
	}

	for _, fhs := range r.MultipartForm.File {
		if len(fhs) == 0 {
			continue
		}

		fh := fhs[0]
		if f, e := fh.Open(); e == nil {
			data, err = ioutil.ReadAll(f)
			f.Close()

			return fh.Filename, data, err
		}
	}

	return "", nil, nil
}
