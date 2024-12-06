package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Error string `json:"error,omitempty"`
		Valid bool   `json:"valid,omitempty"`
	}

	type Parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	code := http.StatusOK

	res := Response{
		Valid: true,
	}

	if len(params.Body) > 140 {
		res = Response{
			Error: "Chirp is too long",
		}

		code = http.StatusBadRequest
	}

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
