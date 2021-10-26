package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

func main() {
	forceCacheTime := flag.Duration("force", 0, "force boltCache time")
	cacheName := flag.String("cache", "", "default for memory cache, non-empty string use boltdb")
	proxyHost := flag.String("proxy", "", "proxy target address, like http://192.168.1.1:8090")
	port := flag.Int("port", 9338, "port")
	flag.Parse()

	var target *url.URL
	var err error

	if *proxyHost != "" {
		if target, err = url.Parse(*proxyHost); err != nil {
			panic(err)
		}
	}

	http.Handle("/", New(target, *cacheName, *forceCacheTime))
	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

type garnish struct {
	c              Cache
	proxy          *httputil.ReverseProxy
	forceCacheTime time.Duration
	target         *url.URL
}

func New(target *url.URL, cacheName string, forceCacheTime time.Duration) *garnish {
	var proxy *httputil.ReverseProxy
	if target != nil {
		proxy = httputil.NewSingleHostReverseProxy(target)
	}

	var cache Cache
	if cacheName == "" {
		cache = newMemCache()
	} else {
		cache = NewCache(cacheName)
	}

	return &garnish{c: cache, proxy: proxy, forceCacheTime: forceCacheTime, target: target}
}

func (g *garnish) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := g.proxy

	if r.Method != http.MethodGet { // only GET requests should be cached
		w.Header().Set(xGarnishCache, "MISS")
		p.ServeHTTP(w, r)
		return
	}

	c := g.c
	u := r.URL.String()

	var page *Page
	if c.Get(u, false, &page) {
		loadCache(w, page)
		return
	}

	w.Header().Set(xGarnishCache, "MISS")
	pw := &responseWriter{proxied: w}
	start := time.Now()
	p.ServeHTTP(pw, r)
	cost := time.Since(start)

	saveCache(w, pw, cost, g.forceCacheTime, c, u)
}

func saveCache(w http.ResponseWriter, pw *responseWriter, cost, forceCacheTime time.Duration, c Cache, urlAddr string) {
	savePage := &Page{Status: pw.statusCode, Headers: w.Header(), ProxyLoadTime: cost.String()}
	contentType := w.Header().Get("Content-Type")
	savePage.IsText = IsContentTypeText(contentType)

	if savePage.IsText {
		savePage.BodyString = string(pw.body)
	} else {
		savePage.BodyBytes = pw.body
	}

	if forceCacheTime > 0 {
		c.Put(urlAddr, savePage, forceCacheTime)
		return
	}

	cc := w.Header().Get(cacheControl)
	if toCache, duration := parseCacheControl(cc); toCache {
		c.Put(urlAddr, savePage, duration)
	}
}

const xGarnishCache = "X-Garnish-Cache"

func loadCache(w http.ResponseWriter, page *Page) {
	w.Header().Set(xGarnishCache, "HIT")
	w.Header().Set("X-Proxy-Load-Time", page.ProxyLoadTime)
	w.WriteHeader(page.Status)
	for k, v := range page.Headers {
		if k != xGarnishCache {
			w.Header().Set(k, v[0])
		}
	}

	if page.IsText {
		_, _ = w.Write([]byte(page.BodyString))
	} else {
		_, _ = w.Write(page.BodyBytes)
	}
}

type Page struct {
	Status        int
	Headers       http.Header
	BodyBytes     []byte
	BodyString    string
	IsText        bool
	ProxyLoadTime string
}

type responseWriter struct {
	proxied    http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseWriter) Header() http.Header { return r.proxied.Header() }

func (r *responseWriter) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)
	return r.proxied.Write(data)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.proxied.WriteHeader(statusCode)
}

const (
	cacheControl = "Cache-Control"
	ccNoCache    = "no-boltCache"
	ccNoStore    = "no-store"
	ccPrivate    = "private"
)

var (
	maxAgeReg       = regexp.MustCompile(`max-age=(\d+)`)
	sharedMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
)

func parseCacheControl(cc string) (cache bool, duration time.Duration) {
	if cc == ccPrivate || cc == ccNoCache || cc == ccNoStore {
		return false, 0
	}

	if cc == "" {
		return false, 0
	}

	directives := strings.Split(cc, ",")
	for _, directive := range directives {
		directive = strings.ToLower(directive)
		age := maxAgeReg.FindStringSubmatch(directive)
		if len(age) > 0 {
			d, err := strconv.Atoi(age[1])
			if err != nil {
				return false, 0
			}

			cache = true
			duration = time.Duration(d) * time.Second
		}

		age = sharedMaxAgeReg.FindStringSubmatch(directive)
		if len(age) > 0 {
			d, err := strconv.Atoi(age[1])
			if err != nil {
				return false, 0
			}

			cache = true
			duration = time.Duration(d) * time.Second
		}
	}

	return
}

func IsContentTypeText(contentType string) bool {
	return strings.Contains(contentType, "/json") ||
		strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "/javascript")
}

type boltCache struct {
	db *bolt.DB
}

type itemRaw struct {
	Value   json.RawMessage
	Timeout time.Duration
	Expire  time.Time
}

type item struct {
	Value   interface{}
	Timeout time.Duration
	Expire  time.Time
}

func (v *itemRaw) Unmarshal(data []byte) error {
	return json.Unmarshal(data, v)
}

func (v item) Marshal() ([]byte, error) {
	return json.Marshal(v)
}

func (v itemRaw) IsValid() bool {
	return v.Timeout <= 0 || v.Expire.After(time.Now())
}

func (c *boltCache) Put(key string, page *Page, timeout time.Duration) (err error) {
	return c.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("boltCache"))
		if bucket == nil {
			var er error
			if bucket, er = tx.CreateBucket([]byte("boltCache")); er != nil {
				return er
			}
		}

		it := item{Value: page, Timeout: timeout, Expire: time.Now().Add(timeout)}
		dat, er := it.Marshal()
		if er != nil {
			return err
		}
		return bucket.Put([]byte(key), dat)
	})
}

func (c *boltCache) Close() error {
	return c.db.Close()
}

func (c *boltCache) Get(key string, force bool, page **Page) bool {
	data, ok, err := c.GetBytes(key, force)
	if err != nil || !ok {
		return false
	}

	if err := json.Unmarshal(data, page); err != nil {
		return false
	}

	return true
}

func (c *boltCache) GetValue(key string) (v []byte) {
	value, _, _ := c.GetBytes(key, false)
	return value
}

func (c *boltCache) GetBytes(key string, force bool) (v []byte, ok bool, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte("boltCache")); bucket != nil {
			if data := bucket.Get([]byte(key)); len(data) > 0 {
				it := itemRaw{}
				if err := it.Unmarshal(data); err != nil {
					return err
				}
				if ok = force || it.IsValid(); ok {
					v = it.Value
				}
			}
		}
		return nil
	})

	return
}

type Cache interface {
	Get(key string, force bool, page **Page) bool
	Put(key string, page *Page, timeout time.Duration) (err error)
}

func NewCache(name string) Cache {
	db, err := bolt.Open(name, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return &boltCache{db: db}
}

type data struct {
	data    *Page
	expires *time.Time
}

func (d *data) shouldBeCleared() bool {
	if d.expires == nil {
		return false
	}

	return time.Now().After(*d.expires)
}

type memCache struct {
	mutex *sync.Mutex
	data  map[string]*data
}

func (c *memCache) Put(key string, page *Page, timeout time.Duration) (err error) {
	d := data{
		data: page,
	}
	if timeout != 0 {
		t := time.Now().Add(timeout)
		d.expires = &t
	}

	c.mutex.Lock()
	c.data[key] = &d
	c.mutex.Unlock()
	return nil
}

func (c *memCache) Get(key string, force bool, page **Page) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if d, ok := c.data[key]; ok {
		if !force && d.shouldBeCleared() {
			return false
		}

		*page = d.data
		return true
	}
	return false
}

func newMemCache() Cache {
	return &memCache{data: map[string]*data{}, mutex: &sync.Mutex{}}
}
