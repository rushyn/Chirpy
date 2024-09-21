package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func refresh(w http.ResponseWriter, req *http.Request){
	tokenStr := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")


	if db.AuthorizNewAccessToken(tokenStr){
		tokenSig := IssueToken(tokenStr, 0)
		
		type AccsessToekn struct{
			Token string `json:"token"`
		}

		token := AccsessToekn{
			Token: tokenSig,
		}

		data, err := json.Marshal(token)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
			


	}else{
		w.WriteHeader(401)
		return
	}
	
}