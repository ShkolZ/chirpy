package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ShkolZ/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	if err != nil {
		log.Fatalln(err)
	}
	mux := http.NewServeMux()
	fileServeHandler := http.FileServer(http.Dir("."))
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		Queries:        dbQueries,
	}

	mux.Handle("/app/", apiCfg.metricsIncMiddleware(http.StripPrefix("/app/", fileServeHandler)))
	mux.HandleFunc("GET /api/healthz", loggingMiddleware(apiCfg.healthzHandler))
	mux.HandleFunc("GET /admin/metrics", loggingMiddleware(apiCfg.metricsHandler))
	mux.HandleFunc("POST /admin/reset", loggingMiddleware(apiCfg.resetHandler))
	mux.HandleFunc("POST /api/validate_chirp", loggingMiddleware(apiCfg.validateChirpHandler))
	mux.HandleFunc("POST /api/users", loggingMiddleware(apiCfg.createUserHandler))
	mux.HandleFunc("POST /api/reset", loggingMiddleware(apiCfg.resetHandler))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
