package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/gorilla/mux"
)

// mock data Struct
type jsonData struct {
	ID         int       `json:"_id"`
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

	// connect to Database
	mongoIP := "mongodb://127.0.0.1:27017"
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, mongoIP)
	if err != nil {
		fmt.Println("MongoDB connection has an error:", err)
	}
	defer cancle()
	saveCollection := client.Database("testdb").Collection("testCollection")

	// count Documents in the Collection for auto increment of the id
	res, err := saveCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Println("error while counting the Documents:", err)
	}
	fmt.Println("result of CountDocuments:", res)
	res++

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

	// save to database
	res2, err := saveCollection.InsertOne(ctx, bson.M{"ID": int(res), "LongURL": parsedData, "ShortURL": shortURL, "CreateDate": timeNow})
	fmt.Println("inserted Doc:", res2.InsertedID)
	if err != nil {
		fmt.Println("save error:", err)
	}
}

func redirectPage(w http.ResponseWriter, r *http.Request) {
	// search in MongoDDB for the given id in the URL and redirect to the longurl behind the id
	// atm it redirects to the url which is typed after the slash
	params := mux.Vars(r)
	mongoIP := "mongodb://127.0.0.1:27017"
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, mongoIP)
	if err != nil {
		fmt.Println("MongoDB connection has an error:", err)
	}
	defer cancle()
	getCollection := client.Database("testdb").Collection("testCollection")

	fmt.Println("get params for redir:", params["id"])
	filter := bson.M{"ShortURL": params["id"]}
	var result bson.M
	err = getCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("Error while findOne:", err)
	}
	long := result["LongURL"]
	http.Redirect(w, r, long.(string), http.StatusMovedPermanently)
}

func main() {
	r := mux.NewRouter()
	// write init mock data into slice
	now := time.Now()
	mockData = append(mockData, jsonData{ID: 1, LongURL: "google.com", ShortURL: "random!", CreateDate: now})
	mockData = append(mockData, jsonData{ID: 2, LongURL: "twitter.com", ShortURL: "random!2", CreateDate: now})

	// create MongoDB connection
	/* mongoIP := "mongodb://localhost:27017"
	ctx, cancle := context.WithTimeout(context.Background(), 2*time.Second)
	client, err := mongo.Connect(ctx, mongoIP)
	if err != nil {
		fmt.Println("MongoDB connection has an error:", err)
	} */

	// Route Handler
	r.HandleFunc("/", mainPage).Methods("GET")
	r.HandleFunc("/{id}", redirectPage).Methods("GET")
	r.HandleFunc("/get/", getHandler).Methods("GET")
	r.HandleFunc("/save/", saveHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
