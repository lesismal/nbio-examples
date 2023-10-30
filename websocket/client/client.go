package main

import (
	"flag"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	request := "hello"
	err = c.WriteMessage(websocket.BinaryMessage, []byte(request))
	if err != nil {
		log.Fatalf("write: %v", err)
		return
	}

	receiveType, response, err := c.ReadMessage()
	if err != nil {
		log.Println("ReadMessage failed:", err)
		return
	}
	if receiveType != websocket.BinaryMessage {
		log.Println("received type != websocket.BinaryMessage")
		return

	}

	if string(response) != request {
		log.Printf("'%v' != '%v'", len(response), len(request))
		return
	}

	log.Println("success echo websocket.BinaryMessage:", request)
}
