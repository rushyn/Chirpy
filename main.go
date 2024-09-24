package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rushyn/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	DB *database.Queries
	PLATFORM string
}






func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits ++
		fmt.Println(cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

var _ = godotenv.Load()
var dbURL = os.Getenv("DB_URL")


var db, _ = sql.Open("postgres", dbURL)


var apiCfg = apiConfig{
	fileserverHits: 0,
	DB: database.New(db),
	PLATFORM: os.Getenv("PLATFORM"),
}



func main() {
	fmt.Println(dbURL)
	const port = "8080"
	mux := http.NewServeMux()
	//mux.Handle("/app/",  http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /metrics", metrics)
	mux.HandleFunc("/admin/reset", reset)
	mux.HandleFunc("POST /api/users", create_user)
	mux.HandleFunc("POST /api/chirps", create_chirp)
	mux.HandleFunc("GET /api/chirps", get_all_chirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", get_chirp_by_id)


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

	if apiCfg.PLATFORM == "dev"{
		apiCfg.fileserverHits = 0
		err := apiCfg.DB.DeleteAllUsers(req.Context())
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(403)
}

func create_user(w http.ResponseWriter, req *http.Request){
	type email struct{
		Email string `json:"email"`
	}

	type JsonUser struct{
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string	`json:"email"`
	}
	
	decoder := json.NewDecoder(req.Body)

	mail := email{}

	err := decoder.Decode(&mail)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}

	NewUser := database.CreateUserParams{
		ID: newUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: mail.Email,
	}

	dbUser, err := apiCfg.DB.CreateUser(req.Context(), NewUser)
	if err != nil{
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}


	data, err := json.Marshal(JsonUser{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)

}



func create_chirp(w http.ResponseWriter, req *http.Request){
	type chirpIn struct{
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type JsonChirp struct{
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string	`json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	
	decoder := json.NewDecoder(req.Body)

	In := chirpIn{}

	err := decoder.Decode(&In)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}

	NewChirp := database.CreateChirpParams{
		ID: newUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body: In.Body,
		UserID: In.UserID,
	}

	dbChirp, err := apiCfg.DB.CreateChirp(req.Context(), NewChirp)
	if err != nil{
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}


	data, err := json.Marshal(JsonChirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:     dbChirp.Body,
		UserID:		dbChirp.UserID,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)

}


func get_all_chirp(w http.ResponseWriter, req *http.Request){
	type JsonChirp struct{
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string	`json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	allChirps, err:= apiCfg.DB.SelectAllChirps(req.Context())
	if err != nil{
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}


	jsonChirp := []JsonChirp{}

	for _, chirp := range allChirps{
		jsonChirp = append(jsonChirp, JsonChirp(chirp))
	}





	data, err := json.Marshal(jsonChirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)


}


func get_chirp_by_id(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")

	type JsonChirp struct{
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string	`json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	chirp, err:= apiCfg.DB.SelectChirp(req.Context(), id)
	if err != nil{
		log.Printf("Error getting uuid: %s", err)
		w.WriteHeader(500)
		return
	}

	jsonChirp := JsonChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}


	data, err := json.Marshal(jsonChirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)

	
}