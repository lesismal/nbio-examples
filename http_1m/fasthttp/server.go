package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	qps   uint64 = 0
	total uint64 = 0
)

func onEcho(ctx *fasthttp.RequestCtx) {
	data := ctx.PostBody()
	if len(data) > 0 {
		ctx.Write(data)
	} else {
		ctx.Write([]byte(time.Now().Format("20060102 15:04:05")))
	}
	atomic.AddUint64(&qps, 1)
}

func main() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic(err)
		}
	}()

	for _, addr := range addrs {
		server := &fasthttp.Server{Handler: onEcho}
		go server.ListenAndServe(addr)
	}

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
