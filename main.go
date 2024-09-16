package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits ++
		fmt.Println(cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

var apiCfg = apiConfig{
	fileserverHits: 0,
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	//mux.Handle("/app/",  http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /metrics", metrics)
	mux.HandleFunc("/reset", reset)


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

func metrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type:", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Hits: %d", apiCfg.fileserverHits)))	
}

func reset(w http.ResponseWriter, req *http.Request) {
	apiCfg.fileserverHits = 0
	w.WriteHeader(200)
}