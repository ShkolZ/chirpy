package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ShkolZ/chirpy/backend/internal/database"
	"github.com/ShkolZ/chirpy/backend/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load("./../.env")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	if err != nil {
		log.Fatalln(err)
	}
	mux := http.NewServeMux()
	fileServeHandler := http.FileServer(http.Dir("."))
	apiCfg := handlers.ApiConfig{
		FileserverHits: atomic.Int32{},
		Queries:        dbQueries,
	}

	//GET Requests
	mux.Handle("/app/", apiCfg.MetricsIncMiddleware(http.StripPrefix("/app/", fileServeHandler)))
	mux.HandleFunc("GET /api/healthz", apiCfg.LoggingMiddleware(apiCfg.HealthzHandler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.LoggingMiddleware(apiCfg.MetricsHandler))
	mux.HandleFunc("GET /api/chirps", apiCfg.LoggingMiddleware(apiCfg.GetChirpsHandler))
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.LoggingMiddleware(apiCfg.GetChirpHandler))

	//POST Requests
	mux.HandleFunc("POST /admin/reset", apiCfg.LoggingMiddleware(apiCfg.ResetHandler))
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.LoggingMiddleware(apiCfg.ValidateChirpHandler))
	mux.HandleFunc("POST /api/users", apiCfg.LoggingMiddleware(apiCfg.CreateUserHandler))
	mux.HandleFunc("POST /api/chirps", apiCfg.LoggingMiddleware(apiCfg.CreateChirpHandler))
	mux.HandleFunc("POST /api/login", apiCfg.LoggingMiddleware(apiCfg.LoginHandler))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
