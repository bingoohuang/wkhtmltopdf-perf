package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/bingoohuang/gg/pkg/flagparse"
	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/wkp"
	"github.com/bingoohuang/wkp/wkhtml"
	"io"
	"log"
	"mime"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func (Config) VersionInfo() string {
	return "wk(a go wrapper for wkhtmltopdf) v1.0.0 2021-07-28 12:49:04"
}

func (c Config) Usage() string {
	return fmt.Sprintf(`Usage of %s:
  -MaxPoolSize value 进程池大小(默认 %d)
  -Listen value 只取前N个检查(默认 :9337)
  -WkVersion value Wk包装版本号, 0/1/1p/2/2p（默认 2)
  -EnableCacheDir 是否开启Wk的缓存目录（默认 false)
  -v 打印版本号后退出`, c.VersionInfo(), runtime.NumCPU()*10)
}

type Config struct {
	Config  string `flag:"c" usage:"yaml config filepath"`
	Init    bool
	Version bool `flag:"v"`

	MaxPoolSize    int
	Listen         string `val:":9337"`
	WkVersion      string `val:"2"`
	EnableCacheDir bool
}

func (c *Config) PostProcess() {
	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = runtime.NumCPU() * 10
	}
}

//go:embed initassets
var initAssets embed.FS

func main() {
	c := &Config{}
	flagparse.Parse(c, flagparse.AutoLoadYaml("c", "wk.yml"))
	ctl.Config{Initing: c.Init, InitFiles: initAssets}.ProcessInit()
	golog.SetupLogrus()
	log.Printf("config: %+v created", c)

	wk := &wkhtml.ToX{MaxPoolSize: c.MaxPoolSize, CacheDir: c.EnableCacheDir}

	assetFileServer := http.FileServer(http.FS(wkp.Assets))
	http.Handle("/a.html", assetFileServer)
	http.Handle("/b.html", assetFileServer)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := toPdf(wk, c.WkVersion, w, r); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	fmt.Println("listening on ", c.Listen)
	panic(http.ListenAndServe(c.Listen, nil))
}

func toPdf(wk *wkhtml.ToX, wkVersion string, w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		return errors.New("no html found")
	}

	extra := r.URL.Query().Get("extra")
	toPdf := switchVersion(wk, wkVersion, r.URL.Query().Get("v"))
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

func switchVersion(wk *wkhtml.ToX, wkVersion, v string) func(htmlURL string, extraArgs string) (pdf []byte, err error) {
	if v == "" {
		v = wkVersion
	}
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
