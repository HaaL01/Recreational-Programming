package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type CacheProxyServer struct {
	cache  *Cache
	origin *url.URL
	client *http.Client
}

func (s *CacheProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + " " + r.URL.String()

	// HIT: we've seen this request before, replay the stored response
	if cached, ok := s.cache.Get(key); ok {
		writeResponse(w, cached, "HIT")
		return
	}

	// MISS: build a new request to the origin with the same path and query
	target := s.origin.ResolveReference(r.URL)
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := s.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	cached := CachedResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
		ExpiresAt:  time.Now().Add(time.Minute * 1),
	}

	// Only cache successful GETs
	if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
		s.cache.Set(key, cached)
	}

	writeResponse(w, cached, "MISS")
}

func writeResponse(w http.ResponseWriter, resp CachedResponse, cacheStatus string) {
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.Header().Set("X-Cache", cacheStatus)
	w.WriteHeader(resp.StatusCode)
	w.Write(resp.Body)
}

// newAdminMux builds the router for admin commands like clearing the cache.
func newAdminMux(cache *Cache) *http.ServeMux {
	admin := http.NewServeMux()
	admin.HandleFunc("POST /clear-cache", func(w http.ResponseWriter, r *http.Request) {
		cache.Clear()
		log.Println("Cache cleared")
		w.Write([]byte("cache cleared\n"))
	})
	return admin
}

// requestClearCache asks an already running proxy to empty its cache.
func requestClearCache(adminAddr string) {
	resp, err := http.Post("http://"+adminAddr+"/clear-cache", "text/plain", nil)
	if err != nil {
		log.Fatal("could not reach the running proxy: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("clearing the cache failed: ", resp.Status)
	}
	fmt.Println("Cache cleared")
}

func main() {
	// Parse Args
	port := flag.String("port", "8080", "Cache-Proxy Server Port")
	origin := flag.String("origin", "", "URL of the server to forward requests to")
	adminPort := flag.String("admin-port", "9090", "Port for admin commands (localhost only)")
	clearCache := flag.Bool("clear-cache", false, "Clear the cache of a running proxy and exit")
	flag.Parse()

	adminAddr := "127.0.0.1:" + *adminPort

	if *clearCache {
		requestClearCache(adminAddr)
		return
	}

	if *origin == "" {
		log.Fatal("--origin is required")
	}

	originURL, err := url.Parse(*origin)
	if err != nil {
		log.Fatal(err)
	}

	server := &CacheProxyServer{
		cache:  NewCache(),
		origin: originURL,
		client: &http.Client{},
	}

	// Admin server: only reachable from this machine
	go func() {
		log.Printf("Admin server listening on %s", adminAddr)
		log.Fatal(http.ListenAndServe(adminAddr, newAdminMux(server.cache)))
	}()

	addr := ":" + *port
	log.Printf("Caching proxy listening on %s, forwarding to %s", addr, originURL)
	log.Fatal(http.ListenAndServe(addr, server))
}
