package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/SzymonJaroslawski/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db             *database.Queries
	platform       string
	secret         string
	fileserverHits atomic.Int32
}

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	platform := os.Getenv("PLATFORM")
	config := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
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

	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handleCreateUser(w, r, cfg)
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handleCreateChirp(w, r, cfg)
	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handleGetAllChirps(w, r, cfg)
	})

	mux.HandleFunc("GET /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		handleGetChirp(w, r, cfg)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, cfg)
	})

	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		handleRefresh(w, r, cfg)
	})

	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		handleRevoke(w, r, cfg)
	})

	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlePutUsers(w, r, cfg)
	})

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
	})

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	log.Printf("Server runing on: http://localhost:%s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}
