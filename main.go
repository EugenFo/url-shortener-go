package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// mock data Struct
type jsonData struct {
	ID       int    `json:"id"`
	LongURL  string `json:"longurl"`
	ShortURL string `json:"shorturl"`
}

var mockData []jsonData

func mainPage() {
	// should use "html/template" package to render the html pages
}

// func to get all the item in the slice. Just for debug purpose
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockData)
}

// atm it just copies data into the slice, give by a request.
// in future it takes the longurl passed from a form, generates a hash and increments the ID
func saveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newData jsonData
	// takes the request and decodes it.
	_ = json.NewDecoder(r.Body).Decode(&newData)

	mockData = append(mockData, newData)

	// returns the saved data, just for debug purpose
	json.NewEncoder(w).Encode(mockData)

	// need to hash the longurl for the shorturl value
}

func main() {
	r := mux.NewRouter()
	// write init mock data into struct
	mockData = append(mockData, jsonData{ID: 1, LongURL: "google.com", ShortURL: "Ad5T2!"})

	http.HandleFunc("/", mainPage)
	r.HandleFunc("/get/", getHandler).Methods("GET")
	r.HandleFunc("/save/", saveHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
