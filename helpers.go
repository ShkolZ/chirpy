package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func cleanInput(i string) string {
	words := strings.Split(i, " ")
	curseWords := []string{"kerfuffle", "sharbert", "fornax"}

	for i := range words {
		word := strings.ToLower(words[i])
		for _, cWord := range curseWords {
			if word == cWord {
				words[i] = "****"
				log.Println("here")
				break
			}
		}
	}

	cleanString := strings.Join(words, " ")
	return cleanString
}

func respondWithError(w http.ResponseWriter, req *http.Request, err string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	data, _ := json.Marshal(errorResponse{Error: err})
	w.Write(data)
}
