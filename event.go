package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)


type Event struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func chirpy_event(w http.ResponseWriter, req *http.Request) {

	polkaKey := strings.TrimPrefix(req.Header.Get("Authorization"), "ApiKey ")

	if polkaKey != apiCfg.polka_key{
		w.WriteHeader(401)
		return		
	}


	decoder := json.NewDecoder(req.Body)
	event := Event{}
	err := decoder.Decode(&event)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if event.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	if db.SetUserRed(event.Data.UserID){
		w.WriteHeader(204)
		return
	}

	w.WriteHeader(404)
}