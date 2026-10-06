package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CacheProxyServer struct {
	// Add fields for cache and proxy configuration if needed

}

type Cache struct {
	// Add fields for cache implementation if needed
}

func main() {
	// Parse Args
	port := flag.String("port", "8080", "Cache-Proxy Server Port")
	origin := flag.String("serverURL", "", "Reverse Proxy target address")
	flag.Parse()

	custom_port := ":" + *port
	origin_url, err := url.Parse(*origin)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(origin_url)

	http.Handle("/", proxy)

	http.ListenAndServe(custom_port, nil)
}
