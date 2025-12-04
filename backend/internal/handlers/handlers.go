package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ShkolZ/chirpy/backend/internal/auth"
	"github.com/ShkolZ/chirpy/backend/internal/database"
	"github.com/ShkolZ/chirpy/backend/internal/helpers"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Queries        *database.Queries
	SecretKey      string
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

func (cfg *ApiConfig) HealthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) ValidateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}

	type validResponse struct {
		CleanBody string `json:"clean_body"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error decoding chirp",
			Code:  400,
		})
		return
	}
	count := 0
	stringToCheck := params.Body
	for range stringToCheck {
		count++
	}
	if count >= 0 && count <= 140 {
		log.Printf("Chirp is valid\n")

		cleanString := helpers.CleanInput(stringToCheck)
		data, _ := json.Marshal(validResponse{CleanBody: cleanString})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)

	} else {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: errors.New("Chirp is too long"),
			Msg:   "Chirp is too long",
			Code:  400,
		})
		return
	}

}

func (cfg *ApiConfig) CreateUserHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Error with decoding",
			Code:  500,
		})
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error hashing password",
			Code:  500,
		})
		return
	}

	user, err := cfg.Queries.CreateUser(req.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
		Password:  hashedPass,
	})
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error creating user",
			Code:  500,
		})
		return
	}
	log.Println("User was Created")

	user.Password = params.Password
	data, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *ApiConfig) CreateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}
	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error decoding",
			Code:  400,
		})
		return
	}
	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some problem with getting token",
			Code:  401,
		})
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.SecretKey)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Not valid jwt",
			Code:  401,
		})
		return
	}

	chirp, err := cfg.Queries.CreateChirp(req.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		Body:      params.Body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userID,
	})
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error creating chirp",
			Code:  500,
		})
		return
	}

	data, _ := json.Marshal(chirp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)

}

func (cfg *ApiConfig) GetChirpsHandler(w http.ResponseWriter, req *http.Request) {

	chirps, err := cfg.Queries.GetChirps(req.Context())
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Weren't able to get chirps",
			Code:  500,
		})
		return
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error marshaling",
			Code:  500,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *ApiConfig) GetChirpHandler(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error parsing",
			Code:  400,
		})
		return
	}
	chirp, err := cfg.Queries.GetChirpById(req.Context(), id)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error getting chirp by id",
			Code:  400,
		})
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error marshalling",
			Code:  500,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type loginResponse struct {
		database.User
		Token string `json:"token"`
	}

	params := reqParams{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Some error decoding",
			Code:  400,
		})
		return
	}

	user, err := cfg.Queries.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Wrong password or email",
			Code:  401,
		})
		return
	}

	isPassword, err := auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Problem with comparing hashes",
			Code:  500,
		})
	}

	if isPassword {
		var tokenString string
		if params.ExpiresInSeconds == 0 {
			tokenString, err = auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(60*time.Minute))
		} else {
			tokenString, err = auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(params.ExpiresInSeconds)*time.Second)
		}
		if err != nil {
			helpers.RespondWithError(w, req, &helpers.ErrorResponse{
				Error: err,
				Msg:   "Some problem with making JWT",
				Code:  500,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		data, _ := json.Marshal(loginResponse{
			User:  user,
			Token: tokenString,
		})
		w.Write(data)
	} else {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: errors.New("Wrong password"),
			Msg:   "Wrong password or email",
			Code:  401,
		})
	}

}
