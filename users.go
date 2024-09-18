package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type email struct{
	Email string `json:"email"`
}



func validate_users(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	email := email{}
	err := decoder.Decode(&email)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := db.CreateUser(email.Email)
	if err != nil{
		fmt.Println(err)
	}

	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

