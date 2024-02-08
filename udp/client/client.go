package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"time"

	"github.com/lesismal/nbio"
)

func main() {
	var (
		addr    = "127.0.0.1:8888"
		request = []byte("hello")
	)

	g := nbio.NewEngine(nbio.Config{})

	done := make(chan int)
	g.OnData(func(c *nbio.Conn, response []byte) {
		log.Println("onData:", string(response))
		if !bytes.Equal(request, response) {
			log.Panic("not equal")
		}
		close(done)

	})

	err := g.Start()
	if err != nil {
		log.Panic(err)
	}
	defer g.Stop()

	c, err := net.Dial("udp", addr)
	if err != nil {
		log.Panic(err)
	}
	nbc, err := g.AddConn(c)
	if err != nil {
		log.Panic(err)
	}

	nbc.Write(request)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		log.Println("timeout")
	case <-done:
		log.Println("success")
	}
}
