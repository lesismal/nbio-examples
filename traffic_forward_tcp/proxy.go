package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/lesismal/nbio"
)

var (
	network    = "tcp"
	proxyAddr  = "127.0.0.1:8080"
	serverAddr = "127.0.0.1:18080"
)

type Session struct {
	mux   sync.Mutex
	Peer  *nbio.Conn
	Cache []byte
}

func main() {
	go server()
	time.AfterFunc(time.Second, client)

	engine := nbio.NewEngine(nbio.Config{
		Network:            network,
		Addrs:              []string{proxyAddr},
		MaxWriteBufferSize: 6 * 1024 * 1024,
	})

	engine.OnOpen(func(srcConn *nbio.Conn) {
		sess := &Session{}
		srcConn.SetSession(sess)
		// engine.DialAsync(network, dstAddr,  func(dstConn *nbio.Conn, err error) {
		engine.DialAsyncTimeout(network, serverAddr, time.Second*3, func(dstConn *nbio.Conn, err error) {
			if err != nil {
				srcConn.Close()
				return
			}

			dstConn.SetSession(&Session{Peer: srcConn})

			sess.mux.Lock()
			defer sess.mux.Unlock()
			sess.Peer = dstConn
			if len(sess.Cache) > 0 {
				sess.Peer.Write(sess.Cache)
				sess.Cache = nil
			}
		})
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		log.Printf("[%v, %v -> %v] onData: %v", c.RemoteAddr().Network(), c.RemoteAddr().String(), c.LocalAddr().String(), len(data))
		sess, _ := c.Session().(*Session)
		if sess == nil {
			sess = &Session{Cache: append([]byte{}, data...)}
			c.SetSession(sess)
		}
		sess.mux.Lock()
		defer sess.mux.Unlock()
		if sess.Peer == nil {
			sess.Cache = append(sess.Cache, data...)
		} else {
			sess.Peer.Write(data)
		}
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	<-make(chan int)
}

func server() {
	ln, err := net.Listen(network, serverAddr)
	if err != nil {
		panic(err)
	}
	for {
		c, err := ln.Accept()
		if err == nil {
			go func(conn net.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						conn.Close()
						return
					}
					// echo
					conn.Write(buf[:n])
				}
			}(c)
		}
	}
}

func client() {
	conn, err := net.Dial(network, proxyAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	wbuf := make([]byte, 512)
	rbuf := make([]byte, 512)
	for i := 0; i < 3; i++ {
		rand.Read(wbuf)
		n, err := conn.Write(wbuf)
		if err != nil || n != len(wbuf) {
			conn.Close()
			return
		}
		n, err = io.ReadFull(conn, rbuf)
		if err != nil || n != len(wbuf) {
			conn.Close()
			return
		}
		if !bytes.Equal(wbuf, rbuf) {
			panic("not equal")
		}
	}
}
