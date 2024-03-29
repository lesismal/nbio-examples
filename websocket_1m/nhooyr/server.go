package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

var (
	qps   uint64 = 0
	total uint64 = 0
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		mt, data, err := c.Read(ctx)
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = c.Write(ctx, mt, data)
		if err != nil {
			log.Println("write:", err)
			break
		}
		atomic.AddUint64(&qps, 1)
	}
}

func serve(addrs []string) {
	for _, v := range addrs {
		go func(addr string) {
			mux := &http.ServeMux{}
			mux.HandleFunc("/ws", echo)
			server := http.Server{
				Addr:    addr,
				Handler: mux,
			}
			fmt.Println("server exit:", server.ListenAndServe())
		}(v)
	}
}

func main() {
	serve(addrs)
	ticker := time.NewTicker(time.Second)
	for i := 1; true; i++ {
		<-ticker.C
		n := atomic.SwapUint64(&qps, 0)
		total += n
		fmt.Printf("running for %v seconds, NumGoroutine: %v, qps: %v, total: %v\n", i, runtime.NumGoroutine(), n, total)
	}
}

var addrs = []string{
	"localhost:28001",
	"localhost:28002",
	"localhost:28003",
	"localhost:28004",
	"localhost:28005",
	"localhost:28006",
	"localhost:28007",
	"localhost:28008",
	"localhost:28009",
	"localhost:28010",

	"localhost:28011",
	"localhost:28012",
	"localhost:28013",
	"localhost:28014",
	"localhost:28015",
	"localhost:28016",
	"localhost:28017",
	"localhost:28018",
	"localhost:28019",
	"localhost:28020",

	"localhost:28021",
	"localhost:28022",
	"localhost:28023",
	"localhost:28024",
	"localhost:28025",
	"localhost:28026",
	"localhost:28027",
	"localhost:28028",
	"localhost:28029",
	"localhost:28030",

	"localhost:28031",
	"localhost:28032",
	"localhost:28033",
	"localhost:28034",
	"localhost:28035",
	"localhost:28036",
	"localhost:28037",
	"localhost:28038",
	"localhost:28039",
	"localhost:28040",

	"localhost:28041",
	"localhost:28042",
	"localhost:28043",
	"localhost:28044",
	"localhost:28045",
	"localhost:28046",
	"localhost:28047",
	"localhost:28048",
	"localhost:28049",
	"localhost:28050",
}
