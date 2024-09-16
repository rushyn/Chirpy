package main

import (
	"log"
	"net/http"
)



func main() {
	const port = "8080"
	mux := http.NewServeMux()
	mux.Handle("/app/",  http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", healthz)


	svr := &http.Server{
		Addr:	":" + port,
		Handler: mux,
	}

	log.Printf("Http server starting on port: %s\n", port)
	log.Fatal(svr.ListenAndServe())
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))	
}