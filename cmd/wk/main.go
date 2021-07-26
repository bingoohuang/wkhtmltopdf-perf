package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bingoohuang/wkp"
	"github.com/bingoohuang/wkp/wkhtml"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.Handle("/assets/", http.FileServer(http.FS(wkp.Assets)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := toPdf(w, r); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	addr := ":9337"
	fmt.Println("listening on ", addr)
	panic(http.ListenAndServe(addr, nil))
}

func toPdf(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		return errors.New("no html found")
	}

	wk := &wkhtml.ToX{}
	toPdf := wk.ToPdf

	switch v := r.URL.Query().Get("v"); v {
	case "0":
		toPdf = wk.ToPdfV0
	case "1p":
		toPdf = wk.ToPdfV1p
	case "1":
		toPdf = wk.ToPdfV1
	case "2":
		toPdf = wk.ToPdfV2
	}

	extra := r.URL.Query().Get("extra")
	pdf, err := toPdf(url, extra)
	if err != nil {
		return err
	}

	fn := time.Now().Format(`20060102150405000`) + ".pdf"
	if r.URL.Query().Get("dl") != "" {
		cd := mime.FormatMediaType("attachment", map[string]string{"filename": fn})
		w.Header().Set("Content-Disposition", cd)
	}
	w.Header().Set("Content-Type", "application/pdf; charset=UTF-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	if _, err := io.Copy(w, bytes.NewReader(pdf)); err != nil {
		return err
	}

	return nil
}

func ParseUploadFile(r *http.Request) (fn string, data []byte, err error) {
	if err := r.ParseMultipartForm(16 /*16 MiB */ << 20); err != nil {
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
