package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lesismal/nbio"
)

var (
	// just for testing, ignoring the issue of TCP packet fragmentation
	echoCount = 3

	addr = "localhost:8000"

	engine *nbio.Engine
)

func client() {
	engine.DialAsync("tcp", addr, func(c *nbio.Conn, err error) {
		if err != nil {
			log.Println("dial error:", err)
			return
		}
		log.Println("[client] on connected:", c.LocalAddr().String(), c.RemoteAddr().String())
		data := []byte("Hello!")
		nwrite, err := c.Write(data)
		log.Printf("[client] on data: %s -> %s: %s, nwrite: %d, err: %v\n", c.LocalAddr().String(), c.RemoteAddr().String(), string(data), nwrite, err)

		// customize client's data handler if needed
		// c.OnData(func(c *nbio.Conn, data []byte) {
		// 	log.Println("received from server:", string(data))
		// 	c.Close()
		// })
	})
}

func isServerSide(addr string) bool {
	return strings.HasSuffix(addr, ":8000")
}

func main() {
	engine = nbio.NewEngine(nbio.Config{
		Network: "tcp",
		Addrs:   []string{addr},
	})

	engine.OnOpen(func(c *nbio.Conn) {
		log.Println("[server] on connected:", c.LocalAddr().String(), c.RemoteAddr().String())
	})
	engine.OnData(func(c *nbio.Conn, data []byte) {
		if isServerSide(c.LocalAddr().String()) {
			nwrite, err := c.Write(append([]byte{}, data...))
			log.Printf("[server] on data: %s -> %s: %s, nwrite: %d, err: %v\n", c.LocalAddr().String(), c.RemoteAddr().String(), string(data), nwrite, err)
		} else {
			// client side, increse echo count and send back to server
			echoCount--
			if echoCount <= 0 {
				c.Close()
				log.Println("[client] closed connection by client self:", c.LocalAddr().String(), c.RemoteAddr().String())
				return
			}
			nwrite, err := c.Write(append([]byte{}, data...))
			log.Printf("[client] on data: %s -> %s: %s, nwrite: %d, err: %v\n", c.LocalAddr().String(), c.RemoteAddr().String(), string(data), nwrite, err)
		}
	})
	engine.OnClose(func(c *nbio.Conn, err error) {
		if isServerSide(c.LocalAddr().String()) {
			log.Println("[server] on close:", c.LocalAddr().String(), c.RemoteAddr().String())
		} else {
			log.Println("[client] on close:", c.LocalAddr().String(), c.RemoteAddr().String())
		}
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	time.AfterFunc(time.Microsecond*100, client)

	<-make(chan int)
}
