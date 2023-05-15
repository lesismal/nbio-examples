package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	proxyServerAddr = "localhost:8888"

	appServerAddr = "ws://localhost:9999/ws"

	proxyServer *nbhttp.Engine
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	// step 1: authenticated with http request infomations

	// step 2: make sure to dial before Upgrade this conn to websocket,
	// because if we Upgrade first, the client may have received the ws handshake response then the client will begin to send data to us.
	// if we receive the data before we have changed the conn's data handler, the conn will parse the ws frames, then, everything goes wrong.
	dialer := &websocket.Dialer{
		Engine:      proxyServer,
		Upgrader:    newUpgrader(),
		DialTimeout: time.Second * 3,
	}
	dstConn, _, err := dialer.Dial(appServerAddr, nil)
	if err != nil {
		log.Printf("Dial failed: %v, %v", appServerAddr, err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	upgrader := newUpgrader()
	srcConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	// step 3: get tcp conns from both src and dst ws conn
	srcNBConn, _ := srcConn.Conn.(*nbio.Conn)
	dstNBConn, _ := dstConn.Conn.(*nbio.Conn)

	// step 4: set sessions and data handler for tcp conns
	srcNBConn.SetSession(dstNBConn)
	dstNBConn.SetSession(srcNBConn)
	srcNBConn.OnData(func(c *nbio.Conn, data []byte) {
		dstNBConn.Write(data)
	})
	dstNBConn.OnData(func(c *nbio.Conn, data []byte) {
		srcNBConn.Write(data)
	})
}

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", onWebsocket)

	proxyServer = nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{proxyServerAddr},
		Handler: mux,
	})

	// close peer
	proxyServer.OnClose(func(c net.Conn, err error) {
		session := c.(*nbio.Conn).Session()
		if session != nil {
			if nbc, ok := session.(*nbio.Conn); ok {
				nbc.Close()
			}
		}
	})

	err := proxyServer.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	proxyServer.Shutdown(ctx)
}
