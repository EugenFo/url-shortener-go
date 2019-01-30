package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// mock data Struct
type jsonData struct {
	ID         int       `json:"id"`
	LongURL    string    `json:"longurl"`
	ShortURL   string    `json:"shorturl"`
	CreateDate time.Time `json:"createDate"`
}

var mockData []jsonData

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("static/index.html")
	tmpl.Execute(w, nil)
}

// func to get all the item in the slice. Just for debug purpose
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockData)
}

// atm it just copies data into the slice, passed by a form.
func saveHandler(w http.ResponseWriter, r *http.Request) {
	parsedData := r.FormValue("longUrlForm")

	// lookup the last ID in the slice and adds 1
	// pseudo "auto increment"
	lastID := mockData[len(mockData)-1].ID + 1
	fmt.Println(strconv.Itoa(lastID))

	// using unix timestamp for hashing
	timeNow := time.Now()
	timeUnix := timeNow.Unix()

	// at first I'm gonna use SHA1 and only the first 4 bits of the hash
	// need to use bijectiv and base62 encoding. best practice

	// converts int64 into string and creates an SHA1 hash
	hash := sha1.Sum([]byte(strconv.FormatInt(timeUnix, 10)))
	shortURL := hex.EncodeToString(hash[:3])
	fmt.Println(shortURL)

	mockData = append(mockData, jsonData{ID: lastID, LongURL: parsedData, ShortURL: shortURL, CreateDate: timeNow})
}

func main() {
	r := mux.NewRouter()
	// write init mock data into slice
	now := time.Now()
	mockData = append(mockData, jsonData{ID: 1, LongURL: "google.com", ShortURL: "random!", CreateDate: now})
	mockData = append(mockData, jsonData{ID: 2, LongURL: "twitter.com", ShortURL: "random!", CreateDate: now})

	// Route Handler
	r.HandleFunc("/", mainPage).Methods("GET")
	r.HandleFunc("/get/", getHandler).Methods("GET")
	r.HandleFunc("/save/", saveHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
