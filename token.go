package main

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueToken(refreshTokenHex string, expire int) string {

	if expire == 0{
		expire = 60*60
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((time.Duration(expire)) * time.Second)),
		Subject: refreshTokenHex,
		})
	
	tokenSig, err := token.SignedString(apiCfg.jwtSecret)
	if err != nil{
		log.Printf("%s\n", err)
		return ""
	}

	return tokenSig
}