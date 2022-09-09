package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8888,
	})
	defer conn.Close()
	if err != nil {
		return
	}

	for i:=0; i<3; i++{
		request := []byte(fmt.Sprintf("hello %d", i))
		response := make([]byte, 1024)
		_, err = conn.Write(request)
		if packLen, remoteAddr, err := conn.ReadFromUDP(response); err == nil {
			log.Println("onData:", string(response[:packLen]))
			if !bytes.Equal(request, response[:packLen]) {
				log.Panic("not equal")
			}
		} else {
			log.Panic("recv msg failed:", remoteAddr, err)
		}
	}

	log.Println("success")
}
