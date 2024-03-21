package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lesismal/nbio"
)

func main() {
	engine := nbio.NewEngine(nbio.Config{
		Network:            "udp",
		Addrs:              []string{"127.0.0.1:8080"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
		UDPReadTimeout:     time.Second,
	})

	// For the same socket connection(same LocalAddr() and RemoteAddr()),
	// all the *nbio.Conn passed to users in OnOpen/OnData/OnClose handler is the same pointer
	engine.OnOpen(func(c *nbio.Conn) {
		log.Printf("onOpen: [%p, %v]", c, c.RemoteAddr().String())
	})
	engine.OnData(func(c *nbio.Conn, data []byte) {
		log.Printf("onData: [%p, %v], %v", c, c.RemoteAddr().String(), string(data))
		c.Write(append([]byte{}, data...))
	})
	engine.OnClose(func(c *nbio.Conn, err error) {
		log.Printf("onClose: [%p, %v], %v", c, c.RemoteAddr().String(), err)
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	<-make(chan int)
}
