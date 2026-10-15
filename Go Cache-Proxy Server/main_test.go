package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// newTestOrigin starts a fake origin server. It counts how many requests
// reach it, so tests can check whether the proxy really used its cache.
func newTestOrigin(t *testing.T) (*httptest.Server, *atomic.Int32) {
	hits := &atomic.Int32{}
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "hello from %s", r.URL.Path)
	}))
	t.Cleanup(origin.Close)
	return origin, hits
}

// newTestProxy starts our caching proxy in front of the given origin.
func newTestProxy(t *testing.T, origin *httptest.Server) *httptest.Server {
	originURL, err := url.Parse(origin.URL)
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewServer(&CacheProxyServer{
		cache:  NewCache(),
		origin: originURL,
		client: &http.Client{},
	})
	t.Cleanup(proxy.Close)
	return proxy
}

// send makes a request and returns the X-Cache header and the body.
func send(t *testing.T, method, url string) (string, string) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.Header.Get("X-Cache"), string(body)
}

func TestMissThenHit(t *testing.T) {
	origin, hits := newTestOrigin(t)
	proxy := newTestProxy(t, origin)

	cacheStatus, body := send(t, http.MethodGet, proxy.URL+"/hello")
	if cacheStatus != "MISS" {
		t.Errorf("first request: got X-Cache %q, want MISS", cacheStatus)
	}
	if body != "hello from /hello" {
		t.Errorf("first request: got body %q", body)
	}

	cacheStatus, body = send(t, http.MethodGet, proxy.URL+"/hello")
	if cacheStatus != "HIT" {
		t.Errorf("second request: got X-Cache %q, want HIT", cacheStatus)
	}
	if body != "hello from /hello" {
		t.Errorf("second request: got body %q", body)
	}

	if hits.Load() != 1 {
		t.Errorf("origin was hit %d times, want 1", hits.Load())
	}
}

func TestDifferentPathsAreCachedSeparately(t *testing.T) {
	origin, _ := newTestOrigin(t)
	proxy := newTestProxy(t, origin)

	send(t, http.MethodGet, proxy.URL+"/a")
	cacheStatus, body := send(t, http.MethodGet, proxy.URL+"/b")

	if cacheStatus != "MISS" {
		t.Errorf("got X-Cache %q, want MISS", cacheStatus)
	}
	if body != "hello from /b" {
		t.Errorf("got body %q, want the response for /b", body)
	}
}

func TestErrorsAreNotCached(t *testing.T) {
	origin, hits := newTestOrigin(t)
	proxy := newTestProxy(t, origin)

	send(t, http.MethodGet, proxy.URL+"/missing")
	cacheStatus, _ := send(t, http.MethodGet, proxy.URL+"/missing")

	if cacheStatus != "MISS" {
		t.Errorf("got X-Cache %q, want MISS", cacheStatus)
	}
	if hits.Load() != 2 {
		t.Errorf("origin was hit %d times, want 2", hits.Load())
	}
}

func TestPostIsNotCached(t *testing.T) {
	origin, hits := newTestOrigin(t)
	proxy := newTestProxy(t, origin)

	send(t, http.MethodPost, proxy.URL+"/orders")
	cacheStatus, _ := send(t, http.MethodPost, proxy.URL+"/orders")

	if cacheStatus != "MISS" {
		t.Errorf("got X-Cache %q, want MISS", cacheStatus)
	}
	if hits.Load() != 2 {
		t.Errorf("origin was hit %d times, want 2", hits.Load())
	}
}

func TestOriginDownReturnsBadGateway(t *testing.T) {
	origin, _ := newTestOrigin(t)
	proxy := newTestProxy(t, origin)
	origin.Close() // the origin is gone before we send anything

	resp, err := http.Get(proxy.URL + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

func TestExpiredEntryIsMiss(t *testing.T) {
	cache := NewCache()
	cache.Set("GET /old", CachedResponse{ExpiresAt: time.Now().Add(-time.Second)})

	if _, ok := cache.Get("GET /old"); ok {
		t.Error("expired entry was returned, want a miss")
	}
}

func TestClear(t *testing.T) {
	cache := NewCache()
	cache.Set("GET /a", CachedResponse{ExpiresAt: time.Now().Add(time.Minute)})

	cache.Clear()

	if _, ok := cache.Get("GET /a"); ok {
		t.Error("entry still in cache after Clear")
	}
}

func TestClearCacheEndpoint(t *testing.T) {
	cache := NewCache()
	cache.Set("GET /a", CachedResponse{ExpiresAt: time.Now().Add(time.Minute)})
	admin := httptest.NewServer(newAdminMux(cache))
	t.Cleanup(admin.Close)

	resp, err := http.Post(admin.URL+"/clear-cache", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want 200", resp.StatusCode)
	}
	if _, ok := cache.Get("GET /a"); ok {
		t.Error("entry still in cache after POST /clear-cache")
	}

	// A GET must not clear the cache
	resp, err = http.Get(admin.URL + "/clear-cache")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET: got status %d, want 405", resp.StatusCode)
	}
}
