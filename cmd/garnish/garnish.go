package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type garnish struct {
	c     *cache
	proxy *httputil.ReverseProxy
}

func New(origin *url.URL) *garnish {
	reverseProxy := httputil.NewSingleHostReverseProxy(origin)
	return &garnish{c: newCache(), proxy: reverseProxy}
}

func (g *garnish) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only GET requests should be cached
	if r.Method != http.MethodGet {
		w.Header().Set("X-Cache", "MISS")
		g.proxy.ServeHTTP(w, r)
		return
	}

	u := r.URL.String()
	cached := g.c.get(u)
	if cached != nil {
		w.Header().Set("X-Cache", "HIT")
		header := http.Header{}
		json.Unmarshal(g.c.get(u+".headers"), &header)
		for k, v := range header {
			w.Header().Set(k, v[0])
		}

		_, _ = w.Write(cached)
		return
	}

	log.Printf("no cache %s\n", u)

	proxyRW := &responseWriter{proxied: w}
	proxyRW.Header().Set("X-Cache", "MISS")
	g.proxy.ServeHTTP(proxyRW, r)

	cc := w.Header().Get(cacheControl)
	toCache, duration := parseCacheControl(cc)
	log.Printf("toCache %t\n", toCache)
	if toCache {
		g.c.store(u, proxyRW.body, duration)

		header, _ := json.Marshal(w.Header())
		g.c.store(u+".headers", header, duration)
	}
}

type responseWriter struct {
	proxied http.ResponseWriter
	body    []byte
}

func (r *responseWriter) Header() http.Header {
	return r.proxied.Header()
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)
	return r.proxied.Write(data)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.proxied.WriteHeader(statusCode)
}
