// Package main imlements a client for movieinfo service
package main

import (
	"context"
	"log"
	"os"
	"time"

	"lab5/movieapi"

	"google.golang.org/grpc"
)

const (
	address      = "localhost:50051"
	defaultTitle = "Pulp fiction"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := movieapi.NewMovieInfoClient(conn)

	// Contact the server and print out its response.
	title := defaultTitle
	if len(os.Args) > 1 {
		title = os.Args[1]
	}
	// Timeout if server doesn't respond
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r1, err1 := c.GetMovieInfo(ctx, &movieapi.MovieRequest{Title: title})
	if err1 != nil {
		log.Fatalf("could not get movie info: %v", err)
	}
	log.Printf("Movie Info for %s %d %s %v", title, r1.GetYear(), r1.GetDirector(), r1.GetCast())
	//add Jurassic Park to the server
	cast := []string{"Sam Neill", "Laura Dern", "Jeff Goldblum", "Richard Attenborough", "Bob Peck"}
	r2, err2 := c.SetMovieInfo(ctx, &movieapi.MovieData{Title: "Jurassic Park", Year: int32(1993), Director: "Steven Spielberg", Cast: cast})
	if err2 != nil {
		log.Fatalf("could not get movie info: %v", err)
	}
	log.Printf("%s", r2.GetStatus())

	//get Jurassic Park from the server
	title = "Jurassic Park"
	r3, err3 := c.GetMovieInfo(ctx, &movieapi.MovieRequest{Title: title})
	if err3 != nil {
		log.Fatalf("could not get movie info: %v", err)
	}
	log.Printf("Movie Info for %s %d %s %v", title, r3.GetYear(), r3.GetDirector(), r3.GetCast())
}
