package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/bingoohuang/wkp/pkg/ss"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGarnish_CacheRequest(t *testing.T) {
	stop := mockServer()
	defer stop()

	expectedXCacheHeaders := []string{"MISS", "HIT"}
	name := ss.RandStr(10)
	defer os.Remove(name)

	g := New(&url.URL{Scheme: "http", Host: "localhost:8088"}, name, 0)

	for _, expectedHeader := range expectedXCacheHeaders {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8088", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		xcache := w.Header().Get(xGarnishCache)
		assert.Equal(t, expectedHeader, xcache)
	}
}

func TestGarnish_NotCacheableMethods(t *testing.T) {
	stop := mockServer()
	defer stop()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodHead, http.MethodDelete, http.MethodTrace}

	name := ss.RandStr(10)
	defer os.Remove(name)

	g := New(&url.URL{Scheme: "http", Host: "localhost:8088"}, name, 0)

	for _, method := range methods {
		t.Run(fmt.Sprintf("method %s", method), func(t *testing.T) {
			req := httptest.NewRequest(method, "http://localhost:8088", nil)
			// the first call
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			xcache := w.Header().Get(xGarnishCache)
			assert.Equal(t, "MISS", xcache)

			// the second call
			w = httptest.NewRecorder()
			g.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			xcache = w.Header().Get(xGarnishCache)
			assert.Equal(t, "MISS", xcache)
		})
	}
}

func BenchmarkGarnish_ServeHTTP(b *testing.B) {
	stop := mockServer()
	defer stop()
	name := ss.RandStr(10)
	defer os.Remove(name)

	g := New(&url.URL{Scheme: "http", Host: "localhost:8088"}, name, 0)
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8088", nil)
	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, req)
	}
}

func mockServer() func() {
	m := http.NewServeMux()
	s := http.Server{Addr: ":8088", Handler: m}
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=100")
		_, _ = w.Write([]byte("OK"))
	})

	go func() {
		_ = s.ListenAndServe()
	}()

	time.Sleep(time.Millisecond * 10)

	return func() {
		panicOnErr(s.Close())
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestJsonParse(t *testing.T) {
	type T struct {
		Name string
	}

	var v *T
	json.Unmarshal([]byte(`{"name": "bingoo"}`), &v)
	fmt.Println(v)
}

func TestParseCacheControl(t *testing.T) {
	headers := map[string]struct {
		givenHeader    string
		shouldBeCached bool
		cacheTime      time.Duration
	}{
		"empty header": {
			givenHeader:    "",
			shouldBeCached: false,
		},
		"no-boltCache": {
			givenHeader:    "no-boltCache",
			shouldBeCached: false,
		},
		"private": {
			givenHeader:    "private",
			shouldBeCached: false,
		},
		"max-age": {
			givenHeader:    "max-age=123",
			shouldBeCached: true,
			cacheTime:      time.Second * 123,
		},
		"max-age uppercase": {
			givenHeader:    "MAX-AGE=123",
			shouldBeCached: true,
			cacheTime:      time.Second * 123,
		},
		"s-max-age": {
			givenHeader:    "s-max-age=123",
			shouldBeCached: true,
			cacheTime:      time.Second * 123,
		},
		"s-max-age overrides max-age": {
			givenHeader:    "s-max-age=123,s-max-age=321",
			shouldBeCached: true,
			cacheTime:      time.Second * 321,
		},
	}

	for name, tcase := range headers {
		t.Run(name, func(t *testing.T) {
			shouldBeCached, cacheTime := parseCacheControl(tcase.givenHeader)
			assert.Equal(t, tcase.shouldBeCached, shouldBeCached)
			assert.Equal(t, tcase.cacheTime, cacheTime)
		})
	}
}
