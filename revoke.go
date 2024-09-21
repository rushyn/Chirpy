package main

import (
	"net/http"
	"strings"
)

func revoke(w http.ResponseWriter, req *http.Request){
	RefreshToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")



	if db.DeauthorizRefreshToken(RefreshToken){
		w.WriteHeader(204)
		return
	}else{
		w.WriteHeader(401)
		return
	}
	
	
}