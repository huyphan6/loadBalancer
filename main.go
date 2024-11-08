package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	// Initialize all servers as healthy for start-up
	for i := range HealthStatus {
		HealthStatus[i] = true
	}

	// Start health check as a go routine in the background
	go RunHealthCheck()

	// Set up the single reverse proxy with dynamic routing
	proxy := httputil.NewSingleHostReverseProxy(nil)
	// Reverse Proxy Struct requires a director function so I made my own custom Director called dynamicDirector
	proxy.Director = dynamicDirector

	// Runs the load balancer server
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(res, req)
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// The Director function is a request transformer. 
// When a request comes to the reverse proxy, 
// Director modifies it so that the request looks as if it was originally intended for the backend server. 
// After the Director function makes its modifications, the proxy forwards the request to the appropriate target server.

// Modifies requests so that they appear intended for the correct backend server before routing.
func dynamicDirector(req *http.Request) {
	// targetURL := getNextServer()
	targetURL := GetNextHealthyServer()
	req.URL.Scheme = targetURL.Scheme
	req.URL.Host = targetURL.Host
	req.URL.Path = targetURL.Path
}