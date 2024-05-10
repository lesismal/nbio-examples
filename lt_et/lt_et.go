package main

import (
	"context"
	"fmt"
	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/logging"
	"time"
)

func LTServer() {
	engine := nbio.NewEngine(nbio.Config{
		Network:                      "tcp",
		Addrs:                        []string{":8899"},
		ReadBufferSize:               2,
		MaxConnReadTimesPerEventLoop: 1,
		MaxWriteBufferSize:           2,
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		fmt.Printf("【LT】read data len: %d, value: %s\n", len(data), string(data))
	})
	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.start failed: %v\n", err)
		return
	}
	defer engine.Stop()
	time.Sleep(3 * time.Second)
	go Client("localhost:8899", "hello")

	<-make(chan int)
}

// ETServer In ETMod
func ETServer() {
	engine := nbio.NewEngine(nbio.Config{
		Network:                      "tcp",
		Addrs:                        []string{":8888"},
		ReadBufferSize:               2,
		MaxConnReadTimesPerEventLoop: 1,
		MaxWriteBufferSize:           2,
		EpollMod:                     nbio.EPOLLET,
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		fmt.Printf("【ET】read data len: %d, value: %s\n", len(data), string(data))
	})
	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.start failed: %v\n", err)
		return
	}
	defer engine.Stop()
	time.Sleep(3 * time.Second)
	go Client("localhost:8888", "world")

	<-make(chan int)
}

func Client(addr string, data string) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	engine := nbio.NewEngine(nbio.Config{})
	done := make(chan int)

	err := engine.Start()
	if err != nil {
		fmt.Printf("start failed: %v\n", err)
	}
	defer engine.Stop()

	c, err := nbio.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	engine.AddConn(c)
	time.Sleep(2 * time.Second)
	c.Write([]byte(data))

	select {
	case <-ctx.Done():
		logging.Error("timeout")
	case <-done:
		logging.Info("success")
	}
}

// The difference of ET and LT
// In LT mode, Kernel will register epitem to wq again after trigger event, but In ET mode,  Kernel will not register
// epitem to wq

// How nbio handle
// when socket data first come, io poller for per conn will only register Read Event(linux call EPOLLIN, bsd call EVFILT_READ)
// When Read event come
// In LT mode, nbio will read MaxConnReadTimesPerEventLoop * ReadBufferSize size。left data will be read in next Read Event
// In ET mode, nbio will change MaxConnReadTimesPerEventLoop to (1<<31 - 1), this indicates nbio will read all the data

// When Write event come and Kernel can't read all the data
// In LT / ET, nbio will both register the Write event(linux call EPOLLOUT, bsd call EVFILT_WRITE)。and when next Write event arrive
// nbio will call flush() to write left data and finally reset the Read Event
func main() {
	logging.SetLevel(logging.LevelDebug)
	fmt.Println("start to handle LT server")
	go LTServer()
	time.Sleep(10 * time.Second)
	fmt.Println("start to handle ET server")
	go ETServer()
	<-make(chan int)
}
