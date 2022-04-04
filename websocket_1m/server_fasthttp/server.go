package main

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/dgrr/fastws"
	"github.com/valyala/fasthttp"
)

var (
	qps   uint64 = 0
	total uint64 = 0
)

func wsHandler(c *fastws.Conn) {
	defer c.Close()
	var err error
	var msg []byte
	var mode fastws.Mode
	for {
		mode, msg, err = c.ReadMessage(msg[:0])
		if err != nil {
			log.Println("read:", err)
			break
		}
		n, err := c.WriteMessage(mode, msg)
		if err != nil || n != len(msg) {
			log.Printf("write: %v, %v, %v", n, len(msg), err)
			break
		}
		atomic.AddUint64(&qps, 1)
	}
}

func serve(addrs []string) {
	for _, v := range addrs {
		go func(addr string) {
			router := fasthttprouter.New()
			router.GET("/ws", fastws.Upgrade(wsHandler))

			server := fasthttp.Server{
				Handler: router.Handler,
			}

			fmt.Println("server exit:", server.ListenAndServe(addr))
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
