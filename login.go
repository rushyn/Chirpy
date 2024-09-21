package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)


type LogInReqest struct{
	Password string `json:"password"`
	Email string `json:"email"`
	Expires_in_seconds int `json:"expires_in_seconds"`
}

type LogInToken struct{
	ID int `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Is_Chirpy_Red bool `json:"is_chirpy_red"`
}

func validate_login(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	LogInReqest := LogInReqest{}
	err := decoder.Decode(&LogInReqest)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := db.GetUser(LogInReqest.Email)
	if err != nil{
		log.Printf("%s\n", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(LogInReqest.Password))
	if err != nil{
		log.Printf("%s\n", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	keyLength := 32
	refreshToken := make([]byte, keyLength)
	_, err = rand.Read(refreshToken)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	refreshTokenHex := hex.EncodeToString(refreshToken)

	db.UpdateRefreshToken(LogInReqest.Email, refreshTokenHex)

	tokenSig := IssueToken(refreshTokenHex, LogInReqest.Expires_in_seconds)


	logInConfirmToken := LogInToken{
		ID: user.ID,
		Email: user.Email,
		Token: tokenSig,
		RefreshToken: refreshTokenHex,
		Is_Chirpy_Red: user.Is_Chirpy_Red,
	}

	data, err := json.Marshal(logInConfirmToken)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

