package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)


type Clames struct{
	Issuer string `json:"Issuer"`
	IssuedAt jwt.NumericDate `json:"IssuedAt"`
	ExpiresAt jwt.NumericDate `json:"ExpiresAt"`
	Subject string `json:"Subject"`
	jwt.RegisteredClaims
}



func update_user(w http.ResponseWriter, req *http.Request) {

	tokenStr := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")

	token, err := jwt.ParseWithClaims(tokenStr, &Clames{}, func(token *jwt.Token) (interface{}, error) {
		return apiCfg.jwtSecret, nil
	})
	if err != nil {
		w.WriteHeader(401)
		return
	} 

	
	RefreshToken, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting RefreshToken from token: %s", err)
		w.WriteHeader(500)
		return
	}

	decoder := json.NewDecoder(req.Body)
	userUpdate := Email{}
	err = decoder.Decode(&userUpdate)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}


	user, err := db.UpdateUser(userUpdate.Email, userUpdate.Password, RefreshToken)
	if err != nil{
		log.Printf("%s\n", err)
		w.WriteHeader(500)
		return
	}

	UpdatedUser := userConfirm{
		ID: user.ID,
		Email: user.Email,
	}

	data, err := json.Marshal(UpdatedUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

