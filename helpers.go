package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type errorResponse struct {
	Error error
	Msg   string
	Code  int
}

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

func respondWithError(w http.ResponseWriter, req *http.Request, err *errorResponse) {
	type res struct {
		Msg string `json:"msg"`
	}

	log.Printf("%v: %v", err.Msg, err.Error)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	data, _ := json.Marshal(resp{
		Msg: err.Msg,
	})
	w.Write(data)
}
