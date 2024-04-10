package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response %v", payload)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Server error: %s", msg)
	} else {
		log.Printf("Client error: %s", msg)
	}
	type errResponse struct {
		Error string `json:"error"`
	}
	responseWithJSON(w, code, errResponse{Error: msg})
}
