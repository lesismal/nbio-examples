package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	clients = flag.Int("clients", 1, "number of clients")
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		time.AfterFunc(time.Second, func() {
			c.WriteMessage(messageType, data)
		})
		log.Println("onEcho:", string(data))
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
	})

	return u
}

func main() {
	flag.Parse()
	engine := nbhttp.NewEngine(nbhttp.Config{})
	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	for i := 0; i < *clients; i++ {
		go func() {
			u := url.URL{Scheme: "ws", Host: "localhost:8888", Path: "/ws"}
			dialer := &websocket.Dialer{
				Engine:      engine,
				Upgrader:    newUpgrader(),
				DialTimeout: time.Second * 3,
			}
			c, res, err := dialer.Dial(u.String(), nil)
			if err != nil {
				if res != nil && res.Body != nil {
					bReason, _ := io.ReadAll(res.Body)
					fmt.Printf("dial failed: %v, reason: %v\n", err, string(bReason))
				} else {
					fmt.Printf("dial failed: %v\n", err)
				}
				return
			}
			c.WriteMessage(websocket.TextMessage, []byte("hello"))
		}()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	engine.Shutdown(ctx)
}
