// Chat is a server that lets clients chat with each other.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client struct {
	channel chan<- string //client is an outgoing message channel data type
	name    string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) //all incoming client messages
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000") //creat a local tcp server
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept() //wait for the new connection
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool) //a map to store the client status
	for {
		select {
		//whenever recieve msg from the message channel, send that message to everyone
		case msg := <-messages:
			for cli := range clients {
				cli.channel <- msg
			}
		//whenever recive msg from entering channel, add that client status to true
		case cli := <-entering:
			clients[cli] = true
			cli.channel <- "people in the channel" //sent that new arrival the people in the channel
			for k := range clients {
				cli.channel <- k.name
			}
		//whenever reieve msg from leaving channel, delete that client from the map
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.channel)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch) //make a gorountine to print the message

	ch <- "Enter your name:"
	input := bufio.NewScanner(conn)
	var name string
	if input.Scan() {
		name = input.Text() //allow the users to input their name
	}
	var cli = client{ch, name}

	ch <- "You are " + cli.name
	messages <- cli.name + " has arrived"
	entering <- cli

	for input.Scan() {
		messages <- cli.name + ": " + input.Text()
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- cli
	messages <- cli.name + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}
