package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// fuzzyPattern wraps and inserts a '%' between every character of `input` to be used during SQL query searches
func fuzzyPattern(input string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}
	return "%" + strings.Join(words, "%") + "%"
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("respondWithJSON - could not encode payload:", err)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errorResponse{Error: message})
}
