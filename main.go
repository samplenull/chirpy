package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,
		`<html>
  	     <body>
  			  <h1>Welcome, Chirpy Admin</h1>
   			 <p>Chirpy has been visited %d times!</p>
 		</body>
		</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits reset to 0")
}

func main() {
	fmt.Println("Server starting...")
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: loggingHandler(mux),
	}

	apiCfg := &apiConfig{}

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))

	fmt.Println("Server is listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpMessage struct {
		Body string `json:"body"`
	}

	msg := chirpMessage{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check if msg length is 140 or less, otherwise return JSON with error string
	if len(msg.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	type returnValid struct {
		Valid bool `json:"valid"`
	}
	respBody := returnValid{
		Valid: true,
	}

	// If the Chirp is valid, respond with a 200 code and this body - valid: true
	respondWithJSON(w, http.StatusOK, respBody)

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnErr struct {
		Error string `json:"error"`
	}
	respBody := returnErr{
		Error: msg,
	}

	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		fmt.Printf("%s %s %d\n", r.Method, r.URL.Path, rec.statusCode)
	})
}
