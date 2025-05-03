package main

import (
	"fmt"
	"log"
	"net/http"
)

// StartHTTPServer starts a simple HTTP server on the specified port
func StartHTTPServer(port string) {
	// Define the root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s...", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}
