package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

func main() {
	db := database{
		dataMap: map[string]dollars{"shoes": 50, "socks": 5},
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(db.help)) //a help function
	mux.Handle("/list", http.HandlerFunc(db.list))
	mux.Handle("/price", http.HandlerFunc(db.price))
	mux.Handle("/add", http.HandlerFunc(db.add))       //create a new item
	mux.Handle("/update", http.HandlerFunc(db.update)) //update the price of an item
	mux.Handle("/delete", http.HandlerFunc(db.delete)) //delet an item
	log.Fatal(http.ListenAndServe(":8000", mux))
}

//the database is now contains a map and a read write lock
type database struct {
	lock    sync.RWMutex
	dataMap map[string]dollars
}

func (db *database) help(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "url input instruction\n")
	fmt.Fprintf(w, "\"/list\" would list the item in the database\n")
	fmt.Fprintf(w, "\"/price\" would give the price of selected item\n")
	fmt.Fprintf(w, "\"/add\" would add a new item to database\n")
	fmt.Fprintf(w, "\"/update\" would update the price of selected item\n")
	fmt.Fprintf(w, "\"/delete\" would delete the selected item\n")
	fmt.Fprintf(w, "example: curl \"http://localhost:8000/update?item=shoes&price=20\"\n")
}

func (db *database) list(w http.ResponseWriter, req *http.Request) {
	db.lock.RLock()         //read lock added
	defer db.lock.RUnlock() //unlock when executated
	for item, price := range db.dataMap {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}

func (db *database) price(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	db.lock.RLock()
	defer db.lock.RUnlock()
	price, ok := db.dataMap[item]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
		return
	}
	fmt.Fprintf(w, "%s\n", price)
}

//add a new item to the database
func (db *database) add(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	db.lock.Lock() //write lock
	defer db.lock.Unlock()
	_, ok := db.dataMap[item]
	if ok { //item is already existed
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%q is already in the list, pick another name or use /update to change the price\n", item)
		return
	}
	price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64) //get the new price of the item, and change to a number
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "invalid price, need a real number") //invalid price
		return
	}
	db.dataMap[item] = dollars(price) //add the item to the list and assign a price
	fmt.Fprintf(w, "%q is successfully add to the list\n", item)
}

//update the price of the item in the list
func (db *database) update(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	db.lock.Lock()
	defer db.lock.Unlock()
	_, ok := db.dataMap[item]
	if !ok { //the item is not in the list
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
		return
	}
	price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "invalid price, need a real number")
		return
	}
	db.dataMap[item] = dollars(price) //update the price of that item
	fmt.Fprintf(w, "price of %q is change to %s\n", item, dollars(price))
}

//delete an item form the database
func (db *database) delete(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	db.lock.Lock()
	defer db.lock.Unlock()
	_, ok := db.dataMap[item]
	if !ok { //the item is not in the list
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
		return
	}
	delete(db.dataMap, item) //delete the item from the database
	fmt.Fprintf(w, "%q is successfully delete from the list\n", item)
}
