package main

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/Park-Jiyeonn/coreRPC/xclient"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"
)

type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Error         string
}

func handleConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	reader := gob.NewDecoder(conn)
	for {
		h := &Header{}
		_ = reader.Decode(&h)
		fmt.Println("收到client发来的数据：", h)
	}
}

func client(addr chan string) {
	conn, err := net.Dial("tcp", <-addr)
	if err != nil {
		fmt.Println("dial failed, err", err)
		return
	}

	buf := bufio.NewWriter(conn)
	dec := gob.NewEncoder(buf)
	defer func() {
		err = buf.Flush()
		if err != nil {
			fmt.Println(err)
			_ = conn.Close()
		}
	}()

	for i := 0; i < 10; i++ {
		msg := &Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		err = dec.Encode(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func server(addr chan string) {
	listen, err := net.Listen("tcp", "127.0.0.1:30000")
	addr <- "127.0.0.1:30000"
	defer func() {
		_ = listen.Close()
	}()
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}
	var cnt int
	for {
		cnt++
		fmt.Println("I, uh...", cnt)
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}
		go handleConn(conn)
	}
}

var d *xclient.GeeRegistryDiscovery
var xc *xclient.XClient

type Treasure struct {
	Value  int `json:"value"`
	Weight int `json:"weight"`
}
type TreasureResponse struct {
	StatusCode int        `json:"status_code"`
	StatusMsg  string     `json:"status_msg"`
	Treasure   []Treasure `json:"treasure"`
	Capacity   int        `json:"capacity"`
	Limit      int        `json:"limit"`
}

func main() {

	defer func() { _ = xc.Close() }()
	// send request & receive response

	r := gin.Default()
	r.POST("/kk", func(c *gin.Context) {
		n := 50
		resp := new(TreasureResponse)
		if d == nil || xc == nil {
			d = xclient.NewGeeRegistryDiscovery("http://localhost:9999/_geerpc_/registry", 0)
			xc = xclient.NewXClient(d, xclient.RandomSelect, nil)
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
		err := xc.Call(ctx, "Treasure.GetTreasure", n, resp)
		if err != nil {
			fmt.Println(err.Error())
			c.String(400, err.Error())
			return
		}
		fmt.Println(resp)
		c.JSON(http.StatusOK, resp)
	})

	r.Run(":10134")
}
