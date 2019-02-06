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
	ID         int       `json:"ID"`
	LongURL    string    `json:"LongURL"`
	ShortURL   string    `json:"ShortURL"`
	CreateDate time.Time `json:"CreateDate"`
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("static/index.html")
	tmpl.Execute(w, nil)
}

// func to get all the item in the slice. Just for debug purpose
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mongoIP := "mongodb://127.0.0.1:27017"
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, mongoIP)
	if err != nil {
		fmt.Println("MongoDB connection has an error:", err)
	}
	defer cancle()
	newCollection := client.Database("testdb").Collection("testCollection")

	cur, err := newCollection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("error while fetching all data", err)
	}
	defer cur.Close(ctx)
	var rst bson.M
	representData := []string{}
	for cur.Next(ctx) {
		err := cur.Decode(&rst)
		if err != nil {
			fmt.Println("error while decoding:", err)
		}
		/* 		for k := range rst {
			fmt.Println("Key:", k, "value:", rst[k])
		} */
		res1, _ := json.Marshal(rst)
		fmt.Println("string:", string(res1))
		representData = append(representData, string(res1))

	}
	fmt.Println("rst:", rst)
	// json.NewEncoder(w).Encode(rst)
	json.NewEncoder(w).Encode(representData)
}

// atm it just copies data into the slice, passed by a form.
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
