package util

import (
	"fmt"
	"github.com/bingoohuang/wkp/pkg/uuid"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

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
