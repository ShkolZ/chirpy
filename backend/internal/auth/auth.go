package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(pass string) (string, error) {
	hash, err := argon2id.CreateHash(pass, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(pass, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(pass, hash)
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := strings.Split(headers.Get("Authorization"), " ")[1]
	if tokenString == "" {
		return "", errors.New("No token provided")
	}
	return tokenString, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	log.Println(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * 24 * time.Hour)),
	})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	strID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := uuid.Parse(strID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func MakeRefreshToken() (string, error) {
	data := make([]byte, 16)
	rand.Read(data)
	test := hex.EncodeToString(data)
	log.Println(test)
	return hex.EncodeToString(data), nil
}
