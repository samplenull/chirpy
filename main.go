package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	// Use the http.NewServeMux's .Handle() method to add a handler for the root path (/).
	// Use a standard http.FileServer as the handler
	mux.Handle("/", http.FileServer(http.Dir(".")))
	server.ListenAndServe()
}
