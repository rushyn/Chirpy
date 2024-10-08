package database

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct{
	ID int `json:"id"`
	Author_ID int `json:"author_id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Password []byte `json:"password"`
	RefreshToken string `json:"refresh_token"`
	ExpirerAt time.Time `json:"expirerat"`
	Is_Chirpy_Red bool `json:"is_chirpy_red"`
}


func NewDB(path string) (*DB, error){

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg{
		fmt.Println("removing database.json")
		err := os.Remove("database.json") 
		if err != nil { 
			fmt.Println(err) 
		} 
	}


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

func (db *DB) CreateChirp(body, refreshToken string) (Chirp, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}


	for _, user := range dbs.Users{
		if user.RefreshToken == refreshToken{
			chirp := Chirp{
				ID : (len(dbs.Chirps) + 1),
				Author_ID: user.ID,
				Body: body,
			}
			dbs.Chirps[len(dbs.Chirps) + 1] = chirp
			db.writeDB(dbs)
			return chirp, nil
		}
	}

	return Chirp{}, errors.New("User Error")
}

func (db *DB) CreateUser(email, password string) (User, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	
	if len(dbs.Users) != 0{
		for _, user := range dbs.Users{
			if user.Email == email{
				return User{}, errors.New("User Exists in database")
			}
		}
	}
	
	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		fmt.Println(err)
	}

	user := User{
		ID : len(dbs.Users) + 1,
		Email: email,
		Password: hashPass,
		Is_Chirpy_Red: false,
	}
	dbs.Users[len(dbs.Users) + 1] = user
	db.writeDB(dbs)

	return user, nil
}

func (db *DB) GetChirps(aurthorID int) ([]Chirp, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	chirps := []Chirp{}
	for k := range dbs.Chirps{
		if dbs.Chirps[k].Author_ID == aurthorID || aurthorID == 0 {
			chirps = append(chirps, dbs.Chirps[k])
		}
	}
	return chirps, nil
}

func (db *DB) GetUser(email string) (User, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	for i := range dbs.Users{
		if dbs.Users[i].Email == email{
			return dbs.Users[i], nil
		}
	}
	return User{}, errors.New("incorrect email or password")
}

func (db *DB) ensureDB() error{
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
	dbs := DBStructure {
		Chirps: make(map[int]Chirp),
		Users: make(map[int]User),
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

func (db *DB) UpdateUserCredentials(newEmail, newPassword, RefreshToken string) (User, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	
	if len(dbs.Users) != 0{
		for i, user := range dbs.Users{
			if user.RefreshToken == RefreshToken{
				user.Email = newEmail
				hashPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil{
					fmt.Println(err)
				}
				user.Password = hashPass
				dbs.Users[i] = user
				db.writeDB(dbs)
				return user, nil
			}
		}
	}

	return User{}, errors.New("something Bad Happend user not found in databse")
}

func (db *DB) UpdateRefreshToken(email, refreshToken string) (User, error){
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	
	if len(dbs.Users) != 0{
		for i, user := range dbs.Users{
			if user.Email == email{
				user.RefreshToken = refreshToken
				user.ExpirerAt = time.Now().Add(24 * 60 * time.Hour)
				dbs.Users[i] = user
				db.writeDB(dbs)
				return user, nil
			}
		}
	}

	return User{}, errors.New("something Bad Happend user not found in databse")
}

func (db *DB) AuthorizNewAccessToken(refreshToken string) bool{
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	if len(dbs.Users) != 0{
		for _, user := range dbs.Users{
			if user.RefreshToken == refreshToken{
				return user.ExpirerAt.After(time.Now())
			}
		}
	}

	return false
}

func (db *DB) DeauthorizRefreshToken(refreshToken string) bool{
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	if len(dbs.Users) != 0{
		for i, user := range dbs.Users{
			if user.RefreshToken == refreshToken{
				user.RefreshToken = ""
				dbs.Users[i] = user
				db.writeDB(dbs)
				return true
			}
		}
	}

	return false
}

func (db *DB) Delete_Chirp(refreshToken string, chirpID int) bool{
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	var userID int = 0

	for _, user := range dbs.Users{
		if user.RefreshToken == refreshToken{
			userID = user.ID
			break
		}
	}

	if userID == 0{
		return false
	}


	for i := range dbs.Chirps{
		if dbs.Chirps[i].ID == chirpID{
			if dbs.Chirps[i].Author_ID == userID {
				delete(dbs.Chirps, i)
				db.writeDB(dbs)
				return true
			}else{
				return false
			}
			
		}
	}

	return false
}


func (db *DB) SetUserRed(userID int) bool {
	dbs, err := db.loadDB()
	if err != nil {
		log.Fatal(err)
	}
	
	for k, user := range dbs.Users{
		if user.ID == userID{
			user.Is_Chirpy_Red = true
			dbs.Users[k] = user
			db.writeDB(dbs)
			return true
		}
	}
	return false
}