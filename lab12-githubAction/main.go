//launch microservice server - main.go
package main

import (
	"CloudNative/lab12/microservice"
	"log"
)

func main() {
	s := microservice.NewServer("", "8000")
	log.Fatal(s.ListenAndServe())
}
