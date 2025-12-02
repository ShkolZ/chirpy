package auth

import (
	"github.com/alexedwards/argon2id"
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
