package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ShkolZ/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	Queries        *database.Queries
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	res := fmt.Sprintf(`
	<html>
		<body>
		<h1>Welcome, Chirpy admin</h1>
		<p>Chirpy has been visited %d times</p>
		</body>
	</html>
	`, cfg.fileserverHits.Load())
	w.Write([]byte(res))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
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

func (cfg *apiConfig) healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}

	type validResponse struct {
		CleanBody string `json:"clean_body"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		errF := fmt.Sprintf("Error decoding parameters: %v\n", err)

		respondWithError(w, req, errF)
		return
	}
	count := 0
	stringToCheck := params.Body
	for range stringToCheck {
		count++
	}
	if count >= 0 && count <= 140 {
		log.Printf("Chirp is valid\n")

		cleanString := cleanInput(stringToCheck)
		data, _ := json.Marshal(validResponse{CleanBody: cleanString})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)

	} else {
		myErr := "The Chirp is too long"
		log.Println(myErr)
		respondWithError(w, req, myErr)
		return
	}

}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email string `json:"email"`
	}
	type userParams struct {
		Id        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&params)

	user, err := cfg.Queries.CreateUser(req.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
	})
	if err != nil {
		errF := fmt.Sprintf("Error while creating user: %v\n", err)

		respondWithError(w, req, errF)
		return
	}
	log.Println("User was Created")
	data, _ := json.Marshal(userParams{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		errF := fmt.Sprintf("Some problem decoding chirp params: %v\n", err)

		respondWithError(w, req, errF)
		return
	}

	chirp, err := cfg.Queries.CreateChirp(req.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		Body:      params.Body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    params.UserId,
	})
	if err != nil {
		errF := fmt.Sprintf("Some problem creating chirp: %v\n", err)
		respondWithError(w, req, errF)
		return
	}

	type chirpParams struct {
		Id        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserId    uuid.UUID `json:"user_id"`
	}

	data, _ := json.Marshal(chirpParams{
		Id:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserId:    chirp.UserID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)

}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.Queries.GetChirps(req.Context())
	if err != nil {
		errF := fmt.Sprintf("Some error retrieving chirps: %v", err)
		respondWithError(w, req, errF)
		return
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		errF := fmt.Sprintf("Some error marshaling json: %v", err)
		respondWithError(w, req, errF)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		errF := fmt.Sprintf("Some problem with parsing id: %v", err)
		respondWithError(w, req, errF)
		return
	}
	chirp, err := cfg.Queries.GetChirpById(req.Context(), id)
	if err != nil {
		errF := fmt.Sprintf("Some problem with retreiving chirp: %v", err)
		respondWithError(w, req, errF)
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		errF := fmt.Sprintf("Some problem with marshaling json: %v", err)
		respondWithError(w, req, errF)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
