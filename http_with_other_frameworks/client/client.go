package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	httpTest("http://localhost:8080/hello")
	websocketTest("localhost:8080")
}

func httpTest(url string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("NewRequest failed: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("client.Do failed: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("read body failed: %v", err)
	}

	log.Printf("http hello success, body: '%v'", string(body))
}

func websocketEcho(c *websocket.Conn, messageType int, data []byte) {
	err := c.WriteMessage(messageType, data)
	if err != nil {
		log.Fatalf("WriteMessage failed: %v", err)
	}

	receiveType, message, err := c.ReadMessage()
	if err != nil {
		log.Fatalf("ReadMessage failed: %v", err)
	}

	if receiveType != messageType {
		log.Fatalf("invalid messageType")
	}

	if !bytes.Equal(data, message) {
		log.Fatalf("invalid data")
	}

	messageTypeString := map[int]string{
		websocket.TextMessage:   "TextMessage",
		websocket.BinaryMessage: "BinaryMessage",
	}
	log.Printf("websocket echo success, messageType: [%v], data: '%v'", messageTypeString[messageType], string(data))
}

func websocketTest(addr string) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer c.Close()

	data := []byte("hello world")
	websocketEcho(c, websocket.TextMessage, data)
	websocketEcho(c, websocket.BinaryMessage, data)
}
