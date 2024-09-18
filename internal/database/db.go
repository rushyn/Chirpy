package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct{
	ID int `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]user `json:"users"`
}

type user struct {
	ID int `json:"id"`
	Email string `json:"email"`
}


func NewDB(path string) (*DB, error){
	db := DB{
		path: path,
	}
	err := db.ensureDB()
	if err != nil{
		log.Fatal(err)
	}
	db.loadDB()


	return &db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error){
	//db.mux.Lock()
	//defer db.mux.Unlock()
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	chirp := Chirp{
		ID : len(dbs.Chirps) + 1,
		Body: body,
	}
	dbs.Chirps[len(dbs.Chirps) + 1] = chirp
	db.writeDB(dbs)

	return chirp, nil
}

func (db *DB) CreateUser(email string) (user, error){
	//db.mux.Lock()
	//defer db.mux.Unlock()
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	user := user{
		ID : len(dbs.Users) + 1,
		Email: email,
	}
	dbs.Users[len(dbs.Users) + 1] = user
	db.writeDB(dbs)

	return user, nil
}

func (db *DB) GetChirps() ([]Chirp, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	chirps := []Chirp{}
	for i := range dbs.Chirps{
		chirps = append(chirps, dbs.Chirps[i])
	}
	return chirps, nil
}

func (db *DB) ensureDB() error{
	//db.mux.Lock()
	//defer db.mux.Unlock()
   	if _, err := os.Stat(db.path); err == nil {
		return nil
   	} else {
		err := os.WriteFile(db.path, []byte(nil), 0666)
		if err != nil {
			return err
		}
   	}
   	return nil
}

func (db *DB) loadDB() (DBStructure, error){
	//db.mux.Lock()
	//defer db.mux.Unlock()
	dbs := DBStructure {
		Chirps: make(map[int]Chirp),
		Users: make(map[int]user),
	}
	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &dbs)
	if err != nil {
		fmt.Println(err)
	}
	return dbs, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	//db.mux.Lock()
	//defer db.mux.Unlock()
	data, err := json.Marshal(dbStructure)
	if err != nil{
		log.Fatal(err)
	}
	err = os.WriteFile(db.path, []byte(data), 0666)
	if err != nil{
		log.Fatal(err)
	}
	return nil
}