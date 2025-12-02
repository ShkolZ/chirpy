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
	//GET Requests
	mux.Handle("/app/", apiCfg.metricsIncMiddleware(http.StripPrefix("/app/", fileServeHandler)))
	mux.HandleFunc("GET /api/healthz", apiCfg.loggingMiddleware(apiCfg.healthzHandler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.loggingMiddleware(apiCfg.metricsHandler))
	mux.HandleFunc("GET /api/chirps", apiCfg.loggingMiddleware(apiCfg.getChirpsHandler))
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.loggingMiddleware(apiCfg.getChirpHandler))

	//POST Requests
	mux.HandleFunc("POST /admin/reset", apiCfg.loggingMiddleware(apiCfg.resetHandler))
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.loggingMiddleware(apiCfg.validateChirpHandler))
	mux.HandleFunc("POST /api/users", apiCfg.loggingMiddleware(apiCfg.createUserHandler))
	mux.HandleFunc("POST /api/chirps", apiCfg.loggingMiddleware(apiCfg.createChirpHandler))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
