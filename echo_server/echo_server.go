package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var useStd = flag.Bool("std", false, "whether using std server")

func onHello(c echo.Context) error {
	return c.String(http.StatusOK, "hello world")
}

func onWebsocket(c echo.Context) error {
	w := c.Response()
	r := c.Request()
	upgrader := websocket.NewUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	wsConn := conn.(*websocket.Conn)
	wsConn.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		c.WriteMessage(messageType, data)
	})
	wsConn.OnClose(func(c *websocket.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
	log.Println("OnOpen:", wsConn.RemoteAddr().String())
	return nil
}

func main() {
	flag.Parse()

	e := echo.New()

	e.GET("/hello", onHello)
	e.GET("/ws", onWebsocket)

	if *useStd {
		serveStd(e)
	} else {
		serveNbio(e)
	}
}

func serveStd(e http.Handler) {
	svr := http.Server{
		Addr:    "localhost:8080",
		Handler: e,
	}
	go svr.ListenAndServe()

	log.Println("serving std")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	svr.Shutdown(ctx)
}

func serveNbio(e http.Handler) {
	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8080"},
	}, e, nil)

	err := svr.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	log.Println("serving nbio")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	svr.Shutdown(ctx)
}
