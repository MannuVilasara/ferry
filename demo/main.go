package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func startServer(port string, name string, crashAfter int) {
	mux := http.NewServeMux()
	var requestCount int32

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If crashAfter is > 0, we track requests and "crash"
		if crashAfter > 0 {
			count := atomic.AddInt32(&requestCount, 1)
			if count > int32(crashAfter) {
				// Simulate the server being dead or hanging
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		log.Printf("[%s] Received request on %s\n", name, r.URL.Path)
		fmt.Fprintf(w, "Hello from %s running on port %s!\n", name, port)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// If we've passed our crash threshold, fail the health check!
		if crashAfter > 0 && atomic.LoadInt32(&requestCount) > int32(crashAfter) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	log.Printf("Starting %s on port %s...\n", name, port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server %s failed: %v", name, err)
	}
}

func main() {
	// API-1 runs forever (0 means never crash)
	go startServer("3001", "API-1", 0)
	
	// API-2 will "crash" (return 500s) after serving exactly 1 request!
	go startServer("3002", "API-2", 1)
	
	// API-3 runs forever
	startServer("3003", "API-3", 0)
}
