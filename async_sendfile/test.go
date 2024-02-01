package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lesismal/nbio/nbhttp"
)

var (
	ioPath     = "/xxx"
	videoPath  = "/mv.mp4"
	fileSize   = 500627803
	defaultBuf = make([]byte, fileSize)

	done = make(chan struct{})
)

func init() {
	// mempool.DefaultMemPool = mempool.NewSTD()
	for i := 0; i < len(defaultBuf); i++ {
		defaultBuf[i] = byte(i) ^ 0x5C
	}
}

func client() {
	defer func() {
		close(done)
	}()

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	reqIO, err := http.NewRequest("Get", "http://localhost:8080"+ioPath, nil)
	if err != nil {
		panic(err)
	}
	reqVideo, err := http.NewRequest("Get", "http://localhost:8080"+videoPath, nil)
	if err != nil {
		panic(err)
	}
	reqIO2, err := http.NewRequest("Get", "http://localhost:8080"+ioPath, nil)
	if err != nil {
		panic(err)
	}
	err = reqIO.Write(conn)
	if err != nil {
		panic(err)
	}
	err = reqVideo.Write(conn)
	if err != nil {
		panic(err)
	}
	err = reqIO2.Write(conn)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(conn)

	time.Sleep(time.Second / 2)
	res, err := http.ReadResponse(reader, reqIO)
	if err != nil {
		panic(err)
	}
	if res.Body != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(body, defaultBuf) {
			panic("not equal")
		}
	}

	res, err = http.ReadResponse(reader, reqVideo)
	if err != nil {
		panic(err)
	}
	if res.Body != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		os.WriteFile("./aaa.mp4", body, 0777)
	}

	res, err = http.ReadResponse(reader, reqIO2)
	if err != nil {
		panic(err)
	}
	if res.Body != nil {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(body, defaultBuf) {
			panic("not equal")
		}
	}
	log.Println("==== success")
}

var fileHandler = http.FileServer(http.Dir("./static"))

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/xxx" {
		w.Write(defaultBuf)
	} else {
		fileHandler.ServeHTTP(w, r)
	}
}

func main() {
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{":8080"},
		Handler: &handler{},
	})
	err := engine.Start()
	if err != nil {
		log.Fatalf("engine.Start failed: %v", err)
	}

	time.AfterFunc(time.Second/10, client)

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
