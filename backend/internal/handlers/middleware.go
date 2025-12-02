package handlers

import (
	"log"
	"net/http"
)

func (cfg *ApiConfig) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL.Path)

		next.ServeHTTP(w, req)
	})
}

func (cfg *ApiConfig) MetricsIncMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}
