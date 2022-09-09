//server.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addrstr := fmt.Sprintf("127.0.0.1:8888")
	addr, err := net.ResolveUDPAddr("udp", addrstr)
	if err != nil {
		fmt.Println("ResolveUDPAddr error:", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		if packLen, remoteAddr, err := conn.ReadFromUDP(buf); err == nil {
			log.Printf("onData: [%p, %v], %v", remoteAddr, remoteAddr.String(), string(buf[:packLen]))
			conn.WriteToUDP(buf[:packLen], remoteAddr)
		} else {
			log.Panic(err)
		}
	}
}
