package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = websocket.NewUpgrader()

func init() {
	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		c.WriteMessage(messageType, data)
	})
	upgrader.OnClose(func(c *websocket.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
}

func FiberOnHello(c fiber.Ctx) error {
	c.SendString("hello world")
	return nil
}

func HTTPOnWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	log.Println("OnOpen:", conn.RemoteAddr().String())
}

func main() {
	router := fiber.New()
	router.Get("/hello", FiberOnHello)
	fiberHandler := adaptor.FiberApp(router)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8080"},

		// For HTTP, use fiber.
		// For Webosocket, use http.Handler;
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/ws") {
				HTTPOnWebsocket(w, r)
			} else {
				fiberHandler(w, r)
			}
		}),
	})

	err := engine.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	log.Println("serving [gin-gonic/gin] on [nbio]")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
