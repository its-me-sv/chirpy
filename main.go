package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/its-me-sv/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	port         = "8080"
	filePathRoot = "."
)

func main() {
	godotenv.Load()

	dbURL, platform, jwtSecret := os.Getenv("DB_URL"), os.Getenv("PLATFORM"), os.Getenv("JWT_SECRET")
	if dbURL == "" {
		log.Fatalln(`missing env variable "DB_URL"`)
	}
	if platform == "" {
		log.Fatalln(`missing env variable "PLATFORM"`)
	}
	if jwtSecret == "" {
		log.Fatalln(`missing env variable "JWT_SECRET"`)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("not able to connect to database, error: %v", err)
	}
	dbQueries := database.New(db)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))

	mux.HandleFunc("GET /api/healthz", handleReadiness)

	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", cfg.handleUserLogin)

	mux.HandleFunc("POST /api/refresh", cfg.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevokeRefreshToken)

	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirpByID)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s\n", port)
	log.Fatalln(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}
