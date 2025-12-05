package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ShkolZ/chirpy/backend/internal/auth"
	"github.com/ShkolZ/chirpy/backend/internal/database"
	"github.com/ShkolZ/chirpy/backend/internal/helpers"
	"github.com/google/uuid"
)

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

func (cfg *ApiConfig) UpdateCredentialsHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "No token in the header",
			Code:  401,
		})
		return
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&params)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldnt decode",
			Code:  401,
		})
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SecretKey)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't validate JWT",
			Code:  401,
		})
		return
	}

	passHash, err := auth.HashPassword(params.Password)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't hash the password",
			Code:  401,
		})
		return
	}

	user, err := cfg.Queries.UpdateCredentials(req.Context(), database.UpdateCredentialsParams{
		ID:       userID,
		Email:    params.Email,
		Password: passHash,
	})
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't update user credentials",
			Code:  401,
		})
		return
	}
	data, err := json.Marshal(user)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Could marshal json",
			Code:  401,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)

}

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type loginResponse struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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
		tokenString, err = auth.MakeJWT(user.ID, cfg.SecretKey)
		if err != nil {
			helpers.RespondWithError(w, req, &helpers.ErrorResponse{
				Error: err,
				Msg:   "Some problem with making JWT",
				Code:  500,
			})
		}

		refToken, err := auth.MakeRefreshToken()
		if err != nil {
			helpers.RespondWithError(w, req, &helpers.ErrorResponse{
				Error: err,
				Msg:   "Couldn make refresh token",
				Code:  500,
			})
			return
		}

		cfg.Queries.CreateRefreshTokenForUser(req.Context(), database.CreateRefreshTokenForUserParams{
			Token:     refToken,
			ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
			RevokedAt: sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user.ID,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		data, _ := json.Marshal(loginResponse{
			User:         user,
			Token:        tokenString,
			RefreshToken: refToken,
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

func (cfg *ApiConfig) RefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't get Bearer token",
			Code:  400,
		})
		return
	}

	dbToken, err := cfg.Queries.GetTokenbyToken(req.Context(), refToken)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Token doesnt exist in db",
			Code:  401,
		})
		return
	}

	if dbToken.RevokedAt.Valid == true || time.Now().After(dbToken.ExpiresAt) {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Non Valid refresh token",
			Code:  401,
		})
		return
	}
	token, err := auth.MakeJWT(dbToken.UserID, cfg.SecretKey)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldnt make jwt",
			Code:  500,
		})
	}
	data, _ := json.Marshal(response{
		Token: token,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *ApiConfig) RevokeRefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	refToken, _ := auth.GetBearerToken(req.Header)
	err := cfg.Queries.RevokeToken(req.Context(), database.RevokeTokenParams{
		Token: refToken,
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Could revoke the token",
			Code:  400,
		})
	}
	w.WriteHeader(204)
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) UserChirpyRedHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldn't decode",
			Code:  500,
		})
		return
	}

	if params.Event != "user.upgrade" {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: fmt.Errorf("event not matching"),
			Msg:   "Wrong event",
			Code:  203,
		})
		return
	}

	err = cfg.Queries.SetChirpyRedTrue(req.Context(), params.Data.UserId)
	if err != nil {
		helpers.RespondWithError(w, req, &helpers.ErrorResponse{
			Error: err,
			Msg:   "Couldnt find a user",
			Code:  404,
		})
		return
	}

	w.WriteHeader(204)

}
