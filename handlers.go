package main

import (
	"fmt"
	"net/http"
)

func handleMetrics(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := `
    <html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
      </body>
    </html>
    `

	fmt.Fprintf(w, html, cfg.fileserverHits.Load())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleReset(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
