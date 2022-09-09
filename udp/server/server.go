package main

import (
	"log"
	"fmt"
	"time"

	"github.com/lesismal/nbio"
)

func main() {
	g := nbio.NewGopher(nbio.Config{
		Network:            "udp",
		Addrs:              []string{"127.0.0.1:8888"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
		UDPReadTimeout: time.Second,
	})

	g.OnData(func(c *nbio.Conn, data []byte) {
		log.Printf("onData: [%p, %v], %v", c, c.RemoteAddr().String(), string(data))
		c.Write(append([]byte{}, data...))
	})
	g.OnClose(func(c *nbio.Conn, err error) {
		log.Printf("onClose: [%v], %v", c.RemoteAddr().String(), err)
	})

	err := g.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer g.Stop()

	g.Wait()
}
