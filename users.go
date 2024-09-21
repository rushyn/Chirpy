package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Email struct{
	Password string `json:"password"`
	Email string `json:"email"`
}

type userConfirm struct{
	ID int `json:"id"`
	Email string `json:"email"`
}


func validate_users(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	email := Email{}
	err := decoder.Decode(&email)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := db.CreateUser(email.Email, email.Password)
	if err != nil{
		log.Printf("%s\n", err)
		w.WriteHeader(500)
		return
	}

	newUser := userConfirm{
		ID: user.ID,
		Email: user.Email,
	}

	data, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

