package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

// KeyValue represents a key value pair
type KeyValue struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	K  string             `bson:"k,omitempty" json:"k",omitempty`
	V  string             `bson:"v,omitempty" json:"v,omitempty"`
}

// Mongo client database connection, as global variable for simplicity
// validate concurrency handling of the go mongo driver
// TODO: validate concorrency safety of the mongo driver
var client *mongo.Client

// collection is a global variable for simplicity
var collection *mongo.Collection

var ( // command line parameters
	databaseName   = flag.String("database", "testing", "database name (default testing)")
	collectionName = flag.String("collection", "keyvalue", "collection name (default keyvalue)")
	drop           = flag.Bool("drop", false, "drop collection (default false)")
	insert         = flag.Int("insert", 0, "insert test records (default 0)")
)

// MongoConnect returns a validated mongo connection
func MongoConnect(ctx context.Context, uri string) (*mongo.Client, error) {
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return client, err
	}

	// Ping Mongo
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return client, err
	}
	return client, nil
}

func main() {
	// Main context, used for cancellation
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// parse command line arguments
	flag.Parse()

	// Connect to mongo
	client, err := MongoConnect(mainCtx, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	// get a handle to the keyvalue collection in testing
	collection = client.Database(*databaseName).Collection(*collectionName)

	// drop the collection
	if *drop {
		err = collection.Drop(mainCtx)
		if err != nil {
			log.Fatal(err)
		}
	}

	// insert test records
	for i := 0; i < *insert; i++ {
		// define a variable of type KeyValue
		var kv = KeyValue{}

		// insert 3 key/value pairs
		for i := 1; i <= 3; i++ {
			kv.K = fmt.Sprintf("key%d", i)
			kv.V = fmt.Sprintf("value%d", i)
			res, err := collection.InsertOne(mainCtx, kv)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Inserted id %s", res.InsertedID)
		}
	}

	// prepare the http server
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/kv", kvGetHandler).Methods("GET")   // kv get handler
	router.HandleFunc("/kv", kvPostHandler).Methods("POST") // kv post handler
	// start the http server
	log.Info("Listening on port 9090")
	log.Fatal(http.ListenAndServe(":9090", router))

}

// kvGetHandler returns all key/value pairs
func kvGetHandler(w http.ResponseWriter, r *http.Request) {
	// get the whole collection
	filter := bson.M{}
	cur, err := collection.Find(nil, filter)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// get all records
	var keyValueArray []KeyValue
	for cur.Next(nil) {
		var kv KeyValue
		err := cur.Decode(&kv)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		keyValueArray = append(keyValueArray, kv)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonString, err := json.Marshal(keyValueArray)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// send out output
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(jsonString))
}

// kvPostHandler is the handler for inserting a key/value pair
func kvPostHandler(w http.ResponseWriter, r *http.Request) {

	//Retrieve body from http request
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Save data into kv struct
	var kv KeyValue
	err = json.Unmarshal(b, &kv)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// insert into mongo collection
	res, err := collection.InsertOne(nil, kv)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// serialize to json
	jsonString, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Set content-type http header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//Send back data as response
	w.Write(jsonString)
}
