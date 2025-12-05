package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ShkolZ/chirpy/backend/internal/auth"
	"github.com/ShkolZ/chirpy/backend/internal/database"
	"github.com/ShkolZ/chirpy/backend/internal/helpers"
	"github.com/google/uuid"
)

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

func (cfg *ApiConfig) DeleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirp_id"))
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "No id provided",
			Code:  400,
		})
		return
	}

	token, _ := auth.GetBearerToken(req.Header)
	userID, _ := auth.ValidateJWT(token, cfg.SecretKey)

	err = cfg.Queries.DeleteChirpById(req.Context(), database.DeleteChirpByIdParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't delete the user",
			Code:  403,
		})
		return
	}

	w.WriteHeader(204)

}
