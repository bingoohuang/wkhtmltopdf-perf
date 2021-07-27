package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	proxyHost := flag.String("proxy", "", "address, or filename which contains address")
	port := flag.Int("port", 9338, "port")
	flag.Parse()

	remote, err := url.Parse(*proxyHost)
	if err != nil {
		panic(err)
	}

	c := newCache()
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("X-Cache", "MISS")
				p.ServeHTTP(w, r)
				return
			}

			r.Host = remote.Host
			u := r.URL.String()

			if cached := c.get(u); cached != nil {
				w.Header().Set("X-Cache", "HIT")
				header := http.Header{}
				json.Unmarshal(c.get(u+".headers"), &header)
				for k, v := range header {
					if k != "X-Cache" {
						w.Header().Set(k, v[0])
					}
				}
				_, _ = w.Write(cached)
				return
			}

			w.Header().Set("X-Cache", "MISS")
			pw := &responseWriter{proxied: w}
			p.ServeHTTP(pw, r)
			c.store(u, pw.body, time.Hour)
			header, _ := json.Marshal(w.Header())
			c.store(u+".headers", header, time.Hour)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy))
	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
