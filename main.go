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

	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	log.Printf("Server runing on: http://localhost:%s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}
