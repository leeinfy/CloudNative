package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://10.1.62.177:27017" //mongodb port
	database        = "webserver"                  //mongodb db
	collection      = "item"                       //mongodb collection
)

var err error
var mongoClient *mongo.Client
var mongoCollection *mongo.Collection
var ctx = context.Background()

type DataCell struct {
	ID    primitive.ObjectID `bson:"_id"`
	Item  string             `bson:"item"`
	Price int                `bson:"price"`
}

func main() {
	//connect mongodb
	//ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	//defer cancel()
	log.Println("start mongodb connection..........")
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	mongoCollection = mongoClient.Database(database).Collection(collection)
	if err != nil {
		log.Fatal(err)
	}
	//start a http server
	log.Println("start http webserver...........")
	mux := http.NewServeMux()
	mux.Handle("/list", http.HandlerFunc(list))
	mux.Handle("/searchItem", http.HandlerFunc(searchItem))
	mux.Handle("/update", http.HandlerFunc(update))
	mux.Handle("/add", http.HandlerFunc(add))
	mux.Handle("/delete", http.HandlerFunc(delete))
	mux.Handle("/searchPrice", http.HandlerFunc(searchPrice))
	log.Fatal(http.ListenAndServe(":8000", mux))
}

func list(w http.ResponseWriter, r *http.Request) {
	log.Println("list request recieved")
	cursor, err := mongoCollection.Find(ctx, bson.M{}) //find everything in the database
	if err != nil {
		log.Println(err)
		return
	}
	var get bson.M
	var cell DataCell
	var bsonBytes []byte
	for cursor.Next(ctx) {
		//decode the each cursor find to a DataCell struct
		err = cursor.Decode(&get)
		if err != nil {
			log.Println(err)
			return
		}
		bsonBytes, err = bson.Marshal(get)
		if err != nil {
			log.Println(err)
			return
		}
		bson.Unmarshal(bsonBytes, &cell)
		fmt.Fprintf(w, "item: %s, price: %d\n", cell.Item, cell.Price)
	}
	log.Println("list request executed")
}

func searchItem(w http.ResponseWriter, req *http.Request) {
	log.Printf("search request recieved")
	item := req.URL.Query().Get("item")                            //Query get the name of the item
	cursor, err := mongoCollection.Find(ctx, bson.M{"item": item}) //find the item in the database
	if err != nil {
		log.Println(err)
		return
	}
	defer cursor.Close(ctx)
	var get bson.M
	var cell DataCell
	var bsonBytes []byte
	if cursor.RemainingBatchLength() != 0 { //check wheather the item is in the database
		for cursor.Next(ctx) {
			err = cursor.Decode(&get)
			if err != nil {
				log.Println(err)
				return
			}
			bsonBytes, err = bson.Marshal(get)
			if err != nil {
				log.Println(err)
				return
			}
			bson.Unmarshal(bsonBytes, &cell)
			fmt.Fprintf(w, "item: %s, price: %d\n", cell.Item, cell.Price)
		}
	} else {
		fmt.Fprintf(w, "%s does not exist", item)
	}
	log.Println("search request executed")
}

func update(w http.ResponseWriter, req *http.Request) {
	log.Println("update request recieved")
	item := req.URL.Query().Get("item")
	price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64)
	if err != nil {
		log.Println(err)
		return
	}
	result, err := mongoCollection.UpdateOne(
		ctx,
		bson.M{"item": item},
		bson.D{
			{"$set", bson.D{{"price", price}}},
		},
	)
	if err != nil {
		log.Println(err)
		return
	} else {
		if result.MatchedCount == 0 {
			fmt.Fprintf(w, "%s does not exist", item)
		} else {
			fmt.Fprintf(w, "price of %s has been updated", item)
		}
	}
	log.Println("update request executed")
}

func add(w http.ResponseWriter, req *http.Request) {
	log.Println("add request recieved")
	item := req.URL.Query().Get("item")
	price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64)
	if err != nil {
		log.Println(err)
		return
	}
	cursor, err := mongoCollection.Find(ctx, bson.M{"item": item})
	if err != nil {
		log.Println(err)
		return
	}
	defer cursor.Close(ctx)
	if cursor.RemainingBatchLength() != 0 {
		fmt.Fprintf(w, "%s already exist", item)
	} else {
		_, err := mongoCollection.InsertOne(
			ctx,
			bson.D{
				{"item", item},
				{"price", price},
			},
		)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprintf(w, "%s is added to collection", item)
	}
	log.Println("add request executed")
}

func delete(w http.ResponseWriter, req *http.Request) {
	log.Println("delete request recieved")
	item := req.URL.Query().Get("item")
	if err != nil {
		log.Println(err)
		return
	}
	result, err := mongoCollection.DeleteOne(
		ctx,
		bson.D{
			{"item", item},
		},
	)
	if err != nil {
		log.Println(err)
		return
	}
	if result.DeletedCount != 0 {
		fmt.Fprintf(w, "%s is deleted from collection", item)
	} else {
		fmt.Fprintf(w, "collection does not have %s", item)
	}

	log.Println("delete request executed")
}

func searchPrice(w http.ResponseWriter, req *http.Request) {
	log.Printf("search request recieved")
	parameter := req.URL.Query().Get("parameter")
	price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64)
	if err != nil {
		log.Println(err)
		return
	}
	cursor, err := mongoCollection.Find(
		ctx,
		bson.D{{"price", bson.D{{parameter, price}}}},
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer cursor.Close(ctx)
	var get bson.M
	var cell DataCell
	var bsonBytes []byte
	if cursor.RemainingBatchLength() != 0 {
		for cursor.Next(ctx) {
			err = cursor.Decode(&get)
			if err != nil {
				log.Println(err)
				return
			}
			bsonBytes, err = bson.Marshal(get)
			if err != nil {
				log.Println(err)
				return
			}
			bson.Unmarshal(bsonBytes, &cell)
			fmt.Fprintf(w, "item: %s, price: %d\n", cell.Item, cell.Price)
		}
	} else {
		fmt.Fprintf(w, "no item found")
	}
	log.Println("price request executed")
}
