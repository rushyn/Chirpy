package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	apiCfg := apiConfig{}
	fsHandler1 := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	//fsHandler2 := apiCfg.middlewareMetricsInc(http.StripPrefix("/app/*", http.FileServer(http.Dir("."))))
	r.Method("GET", "/app/*", fsHandler1)
	r.Method("GET", "/app", fsHandler1)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthz)
	apiRouter.Get("/metrics", apiCfg.metrics)
	apiRouter.Get("/reset", apiCfg.reset)
	r.Mount("/api", apiRouter)
	corsMux := middlewareCors(r)

	//http.ListenAndServe(":8080", corsMux)
	srv := &http.Server{
		Addr:    ":" + "8080",
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", ".", "8080")
	log.Fatal(srv.ListenAndServe())

}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		fmt.Println(cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})

}

func healthz(writer http.ResponseWriter, r *http.Request) {
	writer.WriteHeader(200)
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metrics(writer http.ResponseWriter, r *http.Request) {
	writer.WriteHeader(200)
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(writer http.ResponseWriter, r *http.Request) {
	writer.WriteHeader(200)
	cfg.fileserverHits = 0
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
