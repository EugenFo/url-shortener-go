package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type shortenerData struct {
	ID         int
	ShortURL   string
	LongURL    string
	CreateDate time.Time
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("static/index.html")
	tmpl.Execute(w, nil)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	parsedData := r.FormValue("longUrlForm")

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
	res++

	// using unix timestamp for hashing
	timeNow := time.Now()
	timeUnix := timeNow.Unix()

	// converts int64 into string and creates an SHA1 hash
	hash := sha1.Sum([]byte(strconv.FormatInt(timeUnix, 10)))
	shortURL := hex.EncodeToString(hash[:3])

	// save to database
	_, err = saveCollection.InsertOne(ctx, shortenerData{ID: int(res), ShortURL: shortURL, LongURL: parsedData, CreateDate: timeNow})
	if err != nil {
		fmt.Println("save error:", err)
	}

	type htmlData struct {
		SURL string
		URL  string
	}

	tmpl, err := template.ParseFiles("static/success.html")
	if err != nil {
		fmt.Println("error while processing data for template:", err)
	}
	p := htmlData{SURL: shortURL, URL: r.Host}
	tmpl.Execute(w, p)
}

func redirectPage(w http.ResponseWriter, r *http.Request) {
	// search in MongoDDB for the given id in the URL and redirect to the longurl behind the id
	params := mux.Vars(r)

	// because of double GET request...
	if params["id"] == "favicon.ico" || params["id"] == "" {
		return
	}
	mongoIP := "mongodb://127.0.0.1:27017"
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, mongoIP)
	if err != nil {
		fmt.Println("MongoDB connection has an error:", err)
	}
	defer cancle()
	getCollection := client.Database("testdb").Collection("testCollection")

	filter := bson.D{{"shorturl", params["id"]}}
	var result shortenerData
	err = getCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("Error while findOne:", err)
	}
	http.Redirect(w, r, result.LongURL, http.StatusMovedPermanently)

}

func main() {
	r := mux.NewRouter()

	// Route Handler
	r.HandleFunc("/", mainPage).Methods("GET")
	r.HandleFunc("/{id}", redirectPage).Methods("GET")
	r.HandleFunc("/save/", saveHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
