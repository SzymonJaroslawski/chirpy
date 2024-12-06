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

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
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
