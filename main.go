package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/rushyn/Chirpy/internal/database"
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

var db, _ = database.NewDB("database.json")



func main() {

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg{
		fmt.Println("removing database.json")
		err := os.Remove("database.json") 
		if err != nil { 
			log.Fatal(err) 
		} 
	}

	
	const port = "8080"
	mux := http.NewServeMux()
	//mux.Handle("/app/",  http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", metrics)
	mux.HandleFunc("/api/reset", reset)
	mux.HandleFunc("POST /api/chirps", validate_chirp)
	mux.HandleFunc("GET /api/chirps", get_chirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", get_chirpById)
	mux.HandleFunc("POST /api/users", validate_users)
	


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