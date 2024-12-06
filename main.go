package main

import (
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

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		handleMetrics(w, r, cfg)
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		handleReset(w, r, cfg)
	})

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	log.Printf("Server runing on: http://localhost:%s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}
