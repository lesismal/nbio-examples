package main

import (
	"fmt"

	"github.com/lesismal/nbio"
)

func main() {
	engine := nbio.NewEngine(nbio.Config{
		Network:            "tcp",
		Addrs:              []string{":8888"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		c.Write(append([]byte{}, data...))
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	<-make(chan int)
}
