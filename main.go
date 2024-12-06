package main

import (
	"log"
	"net/http"
)

func main() {
	serve()
}

func serve() {
	const PORT = "8080"

	mux := http.NewServeMux()

	staticDir := http.Dir("./static/")
	fileserverHanlder := http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(staticDir)))

	mux.Handle("/app/", fileserverHanlder)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	log.Printf("Server runing on: http://localhost:%s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}
