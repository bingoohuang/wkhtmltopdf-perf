package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/bingoohuang/wkp"
	"github.com/bingoohuang/wkp/wkhtml"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"
)

func main() {
	wk := &wkhtml.ToX{}
	addr := ":9337"
	flag.IntVar(&wk.MaxPoolSize, "pool-size", 100, "max pool size")
	flag.StringVar(&addr, "listen address", ":9337", "listen address")
	flag.BoolVar(&wk.CacheDir, "cache", false, "enable --cache-dir /tmp/cache-wk/")
	flag.Parse()

	http.Handle("/assets/", http.FileServer(http.FS(wkp.Assets)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := toPdf(wk, w, r); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	fmt.Println("listening on ", addr)
	panic(http.ListenAndServe(addr, nil))
}

func toPdf(wk *wkhtml.ToX, w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		return errors.New("no html found")
	}

	extra := r.URL.Query().Get("extra")
	toPdf := switchVersion(wk, r.URL.Query().Get("v"))
	pdf, err := toPdf(url, extra)
	if err != nil {
		return err
	}

	if r.URL.Query().Get("dl") != "" {
		fn := time.Now().Format(`20060102150405000`) + ".pdf"
		cd := mime.FormatMediaType("attachment", map[string]string{"filename": fn})
		w.Header().Set("Content-Disposition", cd)
	}

	w.Header().Set("Content-Type", "application/pdf; charset=UTF-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	_, err = io.Copy(w, bytes.NewReader(pdf))
	return err
}

func switchVersion(wk *wkhtml.ToX, v string) func(htmlURL string, extraArgs string) (pdf []byte, err error) {
	switch v {
	default:
		return wk.ToPdf
	case "0":
		return wk.ToPdfV0
	case "1p":
		return wk.ToPdfV1p
	case "1":
		return wk.ToPdfV1
	case "2":
		return wk.ToPdfV2
	case "2p":
		return wk.ToPdfV2p
	}
}
