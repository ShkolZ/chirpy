package handlers

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ShkolZ/chirpy/backend/internal/database"

	_ "github.com/lib/pq"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Queries        *database.Queries
	SecretKey      string
}

func (cfg *ApiConfig) HealthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	res := fmt.Sprintf(`
	<html>
		<body>
		<h1>Welcome, Chirpy admin</h1>
		<p>Chirpy has been visited %d times</p>
		</body>
	</html>
	`, cfg.FileserverHits.Load())
	w.Write([]byte(res))
}

func (cfg *ApiConfig) ResetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.FileserverHits = atomic.Int32{}
	platform := os.Getenv("PLATFORM")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Access denied"))
		return
	}
	cfg.Queries.ResetUsers(req.Context())
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
