package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Server starting...")
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/healthz", healthzHandler)

	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", loggingHandler(http.StripPrefix("/app/", fs)))

	fmt.Println("Server is listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Serving %s at %s\n", r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}
