package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	config := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	serve(config)
}

func serve(cfg *apiConfig) {
	const PORT = "8080"

	mux := http.NewServeMux()

	staticDir := http.Dir("./static/")
	fileserverHanlder := http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(staticDir)))

	mux.Handle("/app/", fileserverHanlder)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
	})

	mux.HandleFunc("POST /reset", func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	log.Printf("Server runing on: http://localhost:%s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}
