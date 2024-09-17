package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var profanes = []string{"kerfuffle", "sharbert", "fornax"}

type payload interface {
	message() string
}

type chirpPOST struct {
	Body string `json:"body"`
}


func (c chirpPOST) message() string {
	messageL := strings.Split(strings.ToLower(c.Body), " ")
	messageN := strings.Split(c.Body, " ")
	for i, word := range messageL {
		for _, profane := range profanes {
			if word == profane {
				messageN[i] = "****"
			}
		}
	}
	return strings.Join(messageN, " ")
}

type chirpPOSTcleaned struct {
	Cleaned_body string `json:"cleaned_body"`
}

func (c chirpPOSTcleaned) message() string {
	return c.Cleaned_body
}

type errorJSON struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	errorReturn := errorJSON{
		Error: msg,
	}

	data, err := json.Marshal(errorReturn)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, p payload) {
	chirpCleaned := chirpPOSTcleaned{
		Cleaned_body: p.message(),
	}

	data, err := json.Marshal(chirpCleaned)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func validate_chirp(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	var chirp = chirpPOST{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, chirp)

}
