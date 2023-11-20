package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
	ltls "github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	stdTLSConfig      *tls.Config
	nbTLSConfig       *ltls.Config
	maxBlockingOnline = 10

	engineTransfer *nbhttp.Engine

	muxRouter = &http.ServeMux{}
	ginRouter = gin.New()

	upgrader         = newUpgrader()
	upgraderTransfer = newUpgrader()

	addrMuxSTD              = "localhost:8000"
	addrGinSTD              = "localhost:8001"
	addrMuxIOModMixed       = "localhost:8002"
	addrGinIOModMixed       = "localhost:8003"
	addrMuxIOModBlocking    = "localhost:8004"
	addrGinIOModBlocking    = "localhost:8005"
	addrMuxIOModNonBlocking = "localhost:8006"
	addrGinIOModNonBlocking = "localhost:8007"

	addrMuxSTDTLS              = "localhost:9000"
	addrGinSTDTLS              = "localhost:9001"
	addrMuxIOModMixedTLS       = "localhost:9002"
	addrGinIOModMixedTLS       = "localhost:9003"
	addrMuxIOModBlockingTLS    = "localhost:9004"
	addrGinIOModBlockingTLS    = "localhost:9005"
	addrMuxIOModNonBlockingTLS = "localhost:9006"
	addrGinIOModNonBlockingTLS = "localhost:9007"

	allAddrs []struct {
		tag          string
		addr         string
		ioMod        int
		handler      http.Handler
		stdTLSConfig *tls.Config
		nbTLSConfig  *ltls.Config
	}

	wsUrls = []string{
		"/onWebsocketUpgrade",
		"/onWebsocketUpgradeAndTransferConnToPoller",
		"/onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer",
	}
)

func main() {
	wgShutdown := sync.WaitGroup{}
	wgShutdown.Add(1)
	fnsShutdown := []func(ctx context.Context){
		func(ctx context.Context) {
			defer wgShutdown.Done()
			engineTransfer.Shutdown(ctx)
		},
	}
	for _, config := range allAddrs {
		wgShutdown.Add(1)
		if config.ioMod < 0 {
			s := newHTTPServer(config.tag, config.addr, config.handler, config.stdTLSConfig)
			fnsShutdown = append(fnsShutdown, func(ctx context.Context) {
				defer wgShutdown.Done()
				s.Shutdown(ctx)
			})
		} else {
			s := newNBHTTPServer(config.tag, config.addr, config.handler, config.ioMod, config.nbTLSConfig)
			fnsShutdown = append(fnsShutdown, func(ctx context.Context) {
				defer wgShutdown.Done()
				s.Shutdown(ctx)
			})
		}
	}

	time.Sleep(time.Second)
	for _, config := range allAddrs {
		clients(config.tag, config.addr, ((config.stdTLSConfig != nil) || (config.nbTLSConfig != nil)))
	}
	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)
	// <-interrupt
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, fn := range fnsShutdown {
		go fn(ctx)
	}
	wgShutdown.Wait()
}

func init() {
	cert, err := tls.X509KeyPair(rsaCertPEM, rsaKeyPEM)
	if err != nil {
		log.Fatalf("tls.X509KeyPair failed: %v", err)
	}
	stdTLSConfig = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	lcert, err := ltls.X509KeyPair(rsaCertPEM, rsaKeyPEM)
	if err != nil {
		log.Fatalf("tls.X509KeyPair failed: %v", err)
	}
	nbTLSConfig = &ltls.Config{
		Certificates:       []ltls.Certificate{lcert},
		InsecureSkipVerify: true,
	}

	allAddrs = []struct {
		tag          string
		addr         string
		ioMod        int
		handler      http.Handler
		stdTLSConfig *tls.Config
		nbTLSConfig  *ltls.Config
	}{
		{"addrMuxSTD", addrMuxSTD, -1, muxRouter, nil, nil},
		{"addrGinSTD", addrGinSTD, -1, ginRouter, nil, nil},
		{"addrMuxIOModMixed", addrMuxIOModMixed, nbhttp.IOModMixed, muxRouter, nil, nil},
		{"addrGinIOModMixed", addrGinIOModMixed, nbhttp.IOModMixed, ginRouter, nil, nil},
		{"addrMuxIOModBlocking", addrMuxIOModBlocking, nbhttp.IOModBlocking, muxRouter, nil, nil},
		{"addrGinIOModBlocking", addrGinIOModBlocking, nbhttp.IOModBlocking, ginRouter, nil, nil},
		{"addrMuxIOModNonBlocking", addrMuxIOModNonBlocking, nbhttp.IOModNonBlocking, muxRouter, nil, nil},
		{"addrGinIOModNonBlocking", addrGinIOModNonBlocking, nbhttp.IOModNonBlocking, ginRouter, nil, nil},

		{"addrMuxSTDTLS", addrMuxSTDTLS, -1, muxRouter, stdTLSConfig, nil},
		{"addrGinSTDTLS", addrGinSTDTLS, -1, ginRouter, stdTLSConfig, nil},
		{"addrMuxIOModMixedTLS", addrMuxIOModMixedTLS, nbhttp.IOModMixed, muxRouter, nil, nbTLSConfig},
		{"addrGinIOModMixedTLS", addrGinIOModMixedTLS, nbhttp.IOModMixed, ginRouter, nil, nbTLSConfig},
		{"addrMuxIOModBlockingTLS", addrMuxIOModBlockingTLS, nbhttp.IOModBlocking, muxRouter, nil, nbTLSConfig},
		{"addrGinIOModBlockingTLS", addrGinIOModBlockingTLS, nbhttp.IOModBlocking, ginRouter, nil, nbTLSConfig},
		{"addrMuxIOModNonBlockingTLS", addrMuxIOModNonBlockingTLS, nbhttp.IOModNonBlocking, muxRouter, nil, nbTLSConfig},
		{"addrGinIOModNonBlockingTLS", addrGinIOModNonBlockingTLS, nbhttp.IOModNonBlocking, ginRouter, nil, nbTLSConfig},
	}

	engineTransfer = nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		MaxLoad: 1000000,
	})

	err = engineTransfer.Start()
	if err != nil {
		panic(fmt.Errorf("nbio.Start failed: %v", err))
	}

	upgraderTransfer.Engine = engineTransfer
	muxRouter.HandleFunc("/onHTTP", onHTTP)
	muxRouter.HandleFunc("/onWebsocketUpgrade", onWebsocketUpgrade)
	muxRouter.HandleFunc("/onWebsocketUpgradeAndTransferConnToPoller", onWebsocketUpgradeAndTransferConnToPoller)
	muxRouter.HandleFunc("/onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer", onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer)

	ginRouter.Any("/onHTTP", ginHandler(onHTTP))
	ginRouter.Any("/onWebsocketUpgrade", ginHandler(onWebsocketUpgrade))
	ginRouter.Any("/onWebsocketUpgradeAndTransferConnToPoller", ginHandler(onWebsocketUpgradeAndTransferConnToPoller))
	ginRouter.Any("/onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer", ginHandler(onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer))
}

func ginHandler(h func(http.ResponseWriter, *http.Request)) func(*gin.Context) {
	return func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		h(w, r)
	}
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnOpen(func(c *websocket.Conn) {
		// 	log.Println("OnOpen:", c.RemoteAddr().String())
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		// log.Println("OnMessage:", messageType, string(data))
		c.WriteMessage(messageType, data)
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		// fmt.Println("---------------------------")
		// log.Println("OnClose:", c.RemoteAddr().String(), err)
		// debug.PrintStack()
		// fmt.Println("---------------------------")
	})
	return u
}

func onHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().Format("2006.01.02 15:04:05.999")))
}

func onWebsocketUpgrade(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Upgrade 111:", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	conn.SetSession("Upgrade")
	fmt.Println("Upgrade 222:", conn.RemoteAddr().String())
}

func onWebsocketUpgradeAndTransferConnToPoller(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpgradeAndTransferConnToPoller 111:", r.RemoteAddr)
	conn, err := upgraderTransfer.UpgradeAndTransferConnToPoller(w, r, nil)
	if err != nil {
		panic(err)
	}
	conn.SetSession("UpgradeAndTransferConnToPoller")
	fmt.Println("UpgradeAndTransferConnToPoller 222:", conn.RemoteAddr().String())
}

func onWebsocketUpgradeWithoutHandlingReadForConnFromSTDServer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpgradeWithoutHandlingReadForConnFromSTDServer 111:", r.RemoteAddr)
	conn, err := upgrader.UpgradeWithoutHandlingReadForConnFromSTDServer(w, r, nil)
	if err != nil {
		panic(err)
	}
	conn.SetSession("UpgradeWithoutHandlingReadForConnFromSTDServer")
	go conn.HandleRead(1024 * 4)
	fmt.Println("UpgradeWithoutHandlingReadForConnFromSTDServer 222:", conn.RemoteAddr().String())
}

func newHTTPServer(tag, addr string, handler http.Handler, tlsConfig *tls.Config) *http.Server {
	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
	if tlsConfig == nil {
		go func() {
			log.Printf("http server [%v] exit: %v", tag, server.ListenAndServe())
		}()
	} else {
		go func() {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				panic(err)
			}
			tlsListener := tls.NewListener(ln, tlsConfig)
			log.Printf("http server tls [%v] exit: %v", tag, server.Serve(tlsListener))
		}()
	}

	return server
}

func newNBHTTPServer(tag, addr string, handler http.Handler, ioMod int, tlsConfig *ltls.Config) *nbhttp.Engine {
	addrs := []string{}
	addrsTLS := []string{}
	if tlsConfig == nil {
		addrs = []string{addr}
	} else {
		addrsTLS = []string{addr}
	}

	engine := nbhttp.NewEngine(nbhttp.Config{
		Name:                    tag,
		Network:                 "tcp",
		Addrs:                   addrs,
		AddrsTLS:                addrsTLS,
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		TLSConfig:               tlsConfig,
		Handler:                 handler,
		IOMod:                   ioMod,
		MaxBlockingOnline:       maxBlockingOnline,
	})

	err := engine.Start()
	if err != nil {
		panic(fmt.Errorf("nbio.Start failed: %v", err))
	}
	return engine
}

func httpTest(wg *sync.WaitGroup, tag, addr string, isTLS bool) {
	defer func() {
		time.Sleep(time.Second / 10)
		wg.Done()
	}()

	urlstr := fmt.Sprintf("http://%s/onHTTP", addr)
	if isTLS {
		urlstr = fmt.Sprintf("https://%s/onHTTP", addr)
	}
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		log.Fatalf("[%s] std http client NewRequest failed: %v", tag, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("[%s] std http client client.Do failed: %v", tag, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("[%s] std http client read body failed: %v", tag, err)
	}

	log.Printf("[%s] std http client hello success, body: '%v'", tag, string(body))
}

func nbhttpTest(wg *sync.WaitGroup, tag, addr string, isTLS bool) {
	defer func() {
		time.Sleep(time.Second / 10)
		wg.Done()
	}()

	urlstr := fmt.Sprintf("http://%s/onHTTP", addr)
	if isTLS {
		urlstr = fmt.Sprintf("https://%s/onHTTP", addr)
	}
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		log.Fatalf("[%s] nb http client NewRequest failed: %v", tag, err)
	}

	nbClient := &nbhttp.Client{
		Engine:          engineTransfer,
		Timeout:         time.Second * 3,
		MaxConnsPerHost: 5,
		TLSClientConfig: &ltls.Config{InsecureSkipVerify: true},
	}

	done := make(chan struct{})
	nbClient.Do(req, func(res *http.Response, conn net.Conn, err error) {
		defer func() {
			close(done)
		}()
		if err != nil {
			log.Fatalf("[%s] nb http client.Do failed: %v", tag, err)
			return
		}
		if err == nil && res.Body != nil {
			defer res.Body.Close()
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("[%s] nb http client read body failed: %v", tag, err)
		}

		log.Printf("[%s] nb http client hello success, body: '%v'", tag, string(body))
	})
	<-done
}

var messageTypeString = map[int]string{
	gorillaWs.TextMessage:   "TextMessage",
	gorillaWs.BinaryMessage: "BinaryMessage",
}

func gorillaWebsocketEcho(tag string, c *gorillaWs.Conn, messageType int, data []byte) {
	err := c.WriteMessage(messageType, data)
	if err != nil {
		log.Fatalf("[%s] gorillaWs websocket client WriteMessage failed: %v, %v", tag, c.LocalAddr().String(), err)
	}

	receiveType, message, err := c.ReadMessage()
	if err != nil {
		log.Fatalf("[%s] gorillaWs websocket client ReadMessage failed: %v %v", tag, c.LocalAddr().String(), err)
	}

	if receiveType != messageType {
		log.Fatalf("[%s] gorillaWs websocket client invalid messageType: %v", tag, c.LocalAddr().String())
	}

	if !bytes.Equal(data, message) {
		log.Fatalf("[%s] gorillaWs websocket client invalid data: %v", tag, c.LocalAddr().String())
	}

	log.Printf("[%s] gorillaWs websocket client echo success, messageType: [%v], data: '%v'", tag, messageTypeString[messageType], string(data))
}

func gorillaWebsocketTest(wg *sync.WaitGroup, tag string, addr, path string, isTLS bool) {
	defer func() {
		time.Sleep(time.Second / 10)
		wg.Done()
	}()

	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	if isTLS {
		u.Scheme = "wss"
	}
	gorillaWs.DefaultDialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	c, _, err := gorillaWs.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("[%s] gorillaWs websocket client dial failed: %v", tag, err)
	}
	defer c.Close()

	time.Sleep(time.Second / 5)

	data := []byte("hello world")
	gorillaWebsocketEcho(tag, c, gorillaWs.TextMessage, data)
	gorillaWebsocketEcho(tag, c, gorillaWs.BinaryMessage, data)

	time.Sleep(time.Second / 5)
}

func nbWebsocketTest(wg *sync.WaitGroup, tag string, addr, path string, isTLS bool) {
	defer func() {
		time.Sleep(time.Second / 10)
		wg.Done()
	}()

	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	if isTLS {
		u.Scheme = "wss"
	}

	cnt := 0
	done := make(chan struct{})
	data := []byte("hello world")
	upgrader := websocket.NewUpgrader()
	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, message []byte) {
		switch cnt {
		case 0:
			if messageType != websocket.TextMessage {
				log.Fatalf("[%s] nb websocket client invalid messageType: %v", tag, c.LocalAddr().String())
			}
			if !bytes.Equal(data, message) {
				log.Fatalf("[%s] nb websocket client invalid data: %v", tag, c.LocalAddr().String())
			}
			log.Printf("[%s] nb websocket client  echo success, messageType: [%v], data: '%v'", tag, messageTypeString[int(messageType)], string(data))
		case 1:
			if messageType != websocket.BinaryMessage {
				log.Fatalf("[%s] nb websocket client invalid messageType: %v", tag, c.LocalAddr().String())
			}
			if !bytes.Equal(data, message) {
				log.Fatalf("[%s] nb websocket client invalid data: %v", tag, c.LocalAddr().String())
			}
			log.Printf("[%s] nb websocket client echo success, messageType: [%v], data: '%v'", tag, messageTypeString[int(messageType)], string(data))
			close(done)
		default:
		}
		cnt++
	})
	dialer := &websocket.Dialer{
		Engine:          engineTransfer,
		Upgrader:        upgrader,
		DialTimeout:     time.Second * 5,
		TLSClientConfig: &ltls.Config{InsecureSkipVerify: true},
	}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("[%s] nb websocket client dial failed: %v", tag, err)
		return
	}
	defer c.Close()
	c.WriteMessage(websocket.TextMessage, data)
	c.WriteMessage(websocket.BinaryMessage, data)

	<-done
	time.Sleep(time.Second / 5)
}

func clients(tag, addr string, isTLS bool) {
	fmt.Printf("------------------ %s start ------------------\n", tag)
	defer fmt.Printf("------------------ %s done ------------------\n", tag)

	wg := &sync.WaitGroup{}
	for j := 0; j < maxBlockingOnline*2; j++ {
		wg.Add(1)
		httpTest(wg, tag, addr, isTLS)

		wg.Add(1)
		nbhttpTest(wg, tag, addr, isTLS)

		for k := 0; k < len(wsUrls); k++ {
			wg.Add(1)
			go nbWebsocketTest(wg, tag, addr, wsUrls[k%len(wsUrls)], isTLS)
		}

		for k := 0; k < len(wsUrls); k++ {
			wg.Add(1)
			go gorillaWebsocketTest(wg, tag, addr, wsUrls[k%len(wsUrls)], isTLS)
		}
	}
	wg.Wait()
}

var rsaCertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUJeohtgk8nnt8ofratXJg7kUJsI4wDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMDEyMDcwODIyNThaFw0zMDEy
MDUwODIyNThaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCy+ZrIvwwiZv4bPmvKx/637ltZLwfgh3ouiEaTchGu
IQltthkqINHxFBqqJg44TUGHWthlrq6moQuKnWNjIsEc6wSD1df43NWBLgdxbPP0
x4tAH9pIJU7TQqbznjDBhzRbUjVXBIcn7bNknY2+5t784pPF9H1v7h8GqTWpNH9l
cz/v+snoqm9HC+qlsFLa4A3X9l5v05F1uoBfUALlP6bWyjHAfctpiJkoB9Yw1TJa
gpq7E50kfttwfKNkkAZIbib10HugkMoQJAs2EsGkje98druIl8IXmuvBIF6nZHuM
lt3UIZjS9RwPPLXhRHt1P0mR7BoBcOjiHgtSEs7Wk+j7AgMBAAGjUzBRMB0GA1Ud
DgQWBBQdheJv73XSOhgMQtkwdYPnfO02+TAfBgNVHSMEGDAWgBQdheJv73XSOhgM
QtkwdYPnfO02+TAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBf
SKVNMdmBpD9m53kCrguo9iKQqmhnI0WLkpdWszc/vBgtpOE5ENOfHGAufHZve871
2fzTXrgR0TF6UZWsQOqCm5Oh3URsCdXWewVMKgJ3DCii6QJ0MnhSFt6+xZE9C6Hi
WhcywgdR8t/JXKDam6miohW8Rum/IZo5HK9Jz/R9icKDGumcqoaPj/ONvY4EUwgB
irKKB7YgFogBmCtgi30beLVkXgk0GEcAf19lHHtX2Pv/lh3m34li1C9eBm1ca3kk
M2tcQtm1G89NROEjcG92cg+GX3GiWIjbI0jD1wnVy2LCOXMgOVbKfGfVKISFt0b1
DNn00G8C6ttLoGU2snyk
-----END CERTIFICATE-----
`)

var rsaKeyPEM = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAsvmayL8MImb+Gz5rysf+t+5bWS8H4Id6LohGk3IRriEJbbYZ
KiDR8RQaqiYOOE1Bh1rYZa6upqELip1jYyLBHOsEg9XX+NzVgS4HcWzz9MeLQB/a
SCVO00Km854wwYc0W1I1VwSHJ+2zZJ2Nvube/OKTxfR9b+4fBqk1qTR/ZXM/7/rJ
6KpvRwvqpbBS2uAN1/Zeb9ORdbqAX1AC5T+m1soxwH3LaYiZKAfWMNUyWoKauxOd
JH7bcHyjZJAGSG4m9dB7oJDKECQLNhLBpI3vfHa7iJfCF5rrwSBep2R7jJbd1CGY
0vUcDzy14UR7dT9JkewaAXDo4h4LUhLO1pPo+wIDAQABAoIBAF6yWwekrlL1k7Xu
jTI6J7hCUesaS1yt0iQUzuLtFBXCPS7jjuUPgIXCUWl9wUBhAC8SDjWe+6IGzAiH
xjKKDQuz/iuTVjbDAeTb6exF7b6yZieDswdBVjfJqHR2Wu3LEBTRpo9oQesKhkTS
aFF97rZ3XCD9f/FdWOU5Wr8wm8edFK0zGsZ2N6r57yf1N6ocKlGBLBZ0v1Sc5ShV
1PVAxeephQvwL5DrOgkArnuAzwRXwJQG78L0aldWY2q6xABQZQb5+ml7H/kyytef
i+uGo3jHKepVALHmdpCGr9Yv+yCElup+ekv6cPy8qcmMBqGMISL1i1FEONxLcKWp
GEJi6QECgYEA3ZPGMdUm3f2spdHn3C+/+xskQpz6efiPYpnqFys2TZD7j5OOnpcP
ftNokA5oEgETg9ExJQ8aOCykseDc/abHerYyGw6SQxmDbyBLmkZmp9O3iMv2N8Pb
Nrn9kQKSr6LXZ3gXzlrDvvRoYUlfWuLSxF4b4PYifkA5AfsdiKkj+5sCgYEAzseF
XDTRKHHJnzxZDDdHQcwA0G9agsNj64BGUEjsAGmDiDyqOZnIjDLRt0O2X3oiIE5S
TXySSEiIkxjfErVJMumLaIwqVvlS4pYKdQo1dkM7Jbt8wKRQdleRXOPPN7msoEUk
Ta9ZsftHVUknPqblz9Uthb5h+sRaxIaE1llqDiECgYATS4oHzuL6k9uT+Qpyzymt
qThoIJljQ7TgxjxvVhD9gjGV2CikQM1Vov1JBigj4Toc0XuxGXaUC7cv0kAMSpi2
Y+VLG+K6ux8J70sGHTlVRgeGfxRq2MBfLKUbGplBeDG/zeJs0tSW7VullSkblgL6
nKNa3LQ2QEt2k7KHswryHwKBgENDxk8bY1q7wTHKiNEffk+aFD25q4DUHMH0JWti
fVsY98+upFU+gG2S7oOmREJE0aser0lDl7Zp2fu34IEOdfRY4p+s0O0gB+Vrl5VB
L+j7r9bzaX6lNQN6MvA7ryHahZxRQaD/xLbQHgFRXbHUyvdTyo4yQ1821qwNclLk
HUrhAoGAUtjR3nPFR4TEHlpTSQQovS8QtGTnOi7s7EzzdPWmjHPATrdLhMA0ezPj
Mr+u5TRncZBIzAZtButlh1AHnpN/qO3P0c0Rbdep3XBc/82JWO8qdb5QvAkxga3X
BpA7MNLxiqss+rCbwf3NbWxEMiDQ2zRwVoafVFys7tjmv6t2Xck=
-----END RSA PRIVATE KEY-----
`)
