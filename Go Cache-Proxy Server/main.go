package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
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

func main() {
	// Parse Args
	port := flag.String("port", "8080", "Cache-Proxy Server Port")
	origin := flag.String("origin", "", "URL of the server to forward requests to")
	flag.Parse()

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

	addr := ":" + *port
	log.Printf("Caching proxy listening on %s, forwarding to %s", addr, originURL)
	log.Fatal(http.ListenAndServe(addr, server))
}
