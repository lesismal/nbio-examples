package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8083", "http service address")
var mesgLen = flag.Int("message-len", 1024, "length of message sent")
var clients = flag.Int("clients", 100, "number of clients to simulate")

func main() {
	flag.Parse()

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/wss"}
	log.Printf("connecting to %s", u.String())

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	wg := sync.WaitGroup{}

	wg.Add(*clients)

	for i := 0; i < *clients; i++ {
		go func() {
			defer wg.Done()

			dialer := &websocket.Dialer{
				TLSClientConfig: tlsConfig,
			}

			c, _, err := dialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			defer c.Close()

			text := make([]byte, *mesgLen)
			for i := 0; i < *mesgLen; i++ {
				text[i] = 'A'
			}

			for {
				err := c.WriteMessage(websocket.TextMessage, text)
				if err != nil {
					log.Fatalf("write: %v", err)
					return
				}

				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				if !bytes.Equal(message, text) {
					panic("not equal")
				}
			}
		}()
	}

	wg.Wait()
}
