package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = websocket.NewUpgrader()

func init() {
	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		c.WriteMessage(messageType, data)
	})
}

func onHello(ctx iris.Context) {
	ctx.WriteString("hello world")
}

func onWebsocket(ctx iris.Context) {
	upgrader.OnMessage(func(conn *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		conn.WriteMessage(messageType, data)
	})
	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Println("OnOpen:", conn.RemoteAddr().String())
	})

	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		panic(err)
	}
	conn.OnClose(func(c *websocket.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
}

func main() {
	app := iris.New()
	app.Get("/hello", onHello)
	app.Get("/ws", onWebsocket)

	err := app.Build()
	if err != nil {
		panic(err)
	}

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8080"},
		Handler: app,
	})

	err = engine.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	log.Println("serving [go-chi/chi] on [nbio]")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
