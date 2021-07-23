package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wkperf/wkhtml"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := toPdf(w, r); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	panic(http.ListenAndServe(":9337", nil))
}

func toPdf(w http.ResponseWriter, r *http.Request) error {
	fn, data, err := getUploadFile(r)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	if len(data) == 0 {
		data = []byte(q.Get("html"))
	}
	if len(data) == 0 {
		d := json.NewDecoder(r.Body)
		var body struct {
			Html []byte `json:"html"`
		}
		if err := d.Decode(&body); err == nil {
			data = body.Html
		}
	}
	if len(data) == 0 {
		return errors.New("no html found")
	}

	wk := &wkhtml.ToX{}
	pdf, err := wk.ToPDF(data)
	if err != nil {
		return err
	}

	if fn == "" {
		fn = time.Now().Format(`20060102150405000`)
	}
	if !strings.HasSuffix(fn, ".pdf") {
		fn += ".pdf"
	}

	if q.Get("dl") != "" {
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

func getUploadFile(r *http.Request) (fn string, data []byte, err error) {
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
