package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lesismal/nbio"
)

func main() {
	g := nbio.NewEngine(nbio.Config{
		Network:            "udp",
		Addrs:              []string{"127.0.0.1:8888"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
		UDPReadTimeout:     time.Second,
	})

	g.OnOpen(func(c *nbio.Conn) {
		log.Printf("onOpen: [%p, %v]", c, c.RemoteAddr().String())
	})
	g.OnData(func(c *nbio.Conn, data []byte) {
		log.Printf("onData: [%p, %v], %v", c, c.RemoteAddr().String(), string(data))
		c.Write(append([]byte{}, data...))
	})
	g.OnClose(func(c *nbio.Conn, err error) {
		log.Printf("onClose: [%p, %v], %v", c, c.RemoteAddr().String(), err)
	})

	err := g.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer g.Stop()

	g.Wait()
}
