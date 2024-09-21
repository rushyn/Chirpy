package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/rushyn/Chirpy/internal/database"
)




type apiConfig struct {
	fileserverHits int
	jwtSecret []byte
	polka_key string
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







var db, _ = database.NewDB("database.json")


func main() {


	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}

	data, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SECRET"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	apiCfg.jwtSecret = data
	apiCfg.polka_key = os.Getenv("POLKA_KEY")


	const port = "8080"
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", metrics)
	mux.HandleFunc("/api/reset", reset)
	mux.HandleFunc("POST /api/chirps", validate_chirp)
	mux.HandleFunc("GET /api/chirps", get_chirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", get_chirpById)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", delete_chirpById)
	mux.HandleFunc("POST /api/users", validate_users)
	mux.HandleFunc("POST /api/login", validate_login)
	mux.HandleFunc("PUT /api/users", update_user)
	mux.HandleFunc("POST /api/refresh", refresh)
	mux.HandleFunc("POST /api/revoke", revoke)
	mux.HandleFunc("POST /api/polka/webhooks", chirpy_event)


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
	w.Write([]byte(fmt.Sprintf(	`<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>`, 
apiCfg.fileserverHits)))	
}

func reset(w http.ResponseWriter, req *http.Request) {
	apiCfg.fileserverHits = 0
	w.WriteHeader(200)
}

func get_chirpById(w http.ResponseWriter, req *http.Request) {
	chirpList, err := db.GetChirps()
	if err != nil{
		log.Fatal(err)
	}
	sort.Slice(chirpList, func(i, j int) bool {return chirpList[i].ID < chirpList[j].ID})


	chirpID, err := strconv.Atoi(req.PathValue("chirpID"))
	if chirpID > len(chirpList) || chirpID == 0 || err != nil{
		w.WriteHeader(404)
		return
	}

	data, err := json.Marshal(chirpList[chirpID - 1])
	if err != nil{
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)	
}


func get_chirps(w http.ResponseWriter, req *http.Request) {

	chirpList, err := db.GetChirps()
	if err != nil{
		log.Fatal(err)
	}
	sort.Slice(chirpList, func(i, j int) bool {return chirpList[i].ID < chirpList[j].ID})

	data, err := json.Marshal(chirpList)
	if err != nil{
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)	
}



func validate_access_token(w http.ResponseWriter, req *http.Request) (jwt.Token, error){
	tokenStr := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")

	token, err := jwt.ParseWithClaims(tokenStr, &Clames{}, func(token *jwt.Token) (interface{}, error) {
		return apiCfg.jwtSecret, nil
	})
	if err != nil {
		w.WriteHeader(401)
		return *token, errors.New("token not valid")
	}

	return *token, nil
}


func delete_chirpById(w http.ResponseWriter, req *http.Request) {

	token, err := validate_access_token(w, req)
	if err != nil{
		return
	}

	RefreshToken, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting RefreshToken from token: %s", err)
		w.WriteHeader(500)
		return
	}

	chirpID, _ := strconv.Atoi(req.PathValue("chirpID"))

	if db.Delete_Chirp(RefreshToken, chirpID){
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(204)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(403)

}