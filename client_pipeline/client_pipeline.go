package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	for i := 0; i < 1000; i++ {
		go one()
	}
	<-make(chan int)
}

func one() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	reqData := []byte(fmt.Sprintf("GET /hello HTTP/1.1\r\nHost: localhost:8080\r\nContent-Length: 0\r\nAccept-Encoding: gzip\r\n\r\n"))
	multiReqData := []byte{}
	for i := 0; i < 10; i++ {
		multiReqData = append(multiReqData, reqData...)
	}

	go func() {
		buf := make([]byte, 1024)
		for i:=0; true; i++ {
			_, err := conn.Read(buf)
			if err != nil {
				panic(err)
			}
			// println(i, "data:" ,string(buf[:n]))
		}
	}()

	for {
		n, err := conn.Write(multiReqData)
		if err != nil || n < len(reqData) {
			panic(err)
		}
		time.Sleep(time.Second / 1)
	}

}
