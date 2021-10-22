package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	idIndex    = uint32(0)
	wsUpgrader *websocket.Upgrader
)

type ConnectionContext struct {
	id uint32
}

func onMessage(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
	connCtx := c.Session().(*ConnectionContext)
	fmt.Println("OnMessage:", connCtx.id, c.RemoteAddr().String())
	c.WriteMessage(messageType, data)
}
func onClose(c *websocket.Conn, err error) {
	connCtx := c.Session().(*ConnectionContext)
	fmt.Println("OnClose:", connCtx.id, c.RemoteAddr().String(), err)
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnMessage(onMessage)
	u.OnClose(onClose)
	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	wsConn := conn.(*websocket.Conn)
	wsConn.SetReadDeadline(time.Time{})
	connCtx := &ConnectionContext{atomic.AddUint32(&idIndex, 1)}
	fmt.Println("OnOpen:", connCtx.id, wsConn.RemoteAddr().String())
	wsConn.SetSession(connCtx)
}

func main() {
	flag.Parse()
	wsUpgrader = newUpgrader()
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", onWebsocket)

	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:8888"},
		Handler: mux,
	})

	err := svr.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	svr.Shutdown(ctx)
}
