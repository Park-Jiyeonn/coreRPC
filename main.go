package main

import (
	"core-RPC/codec"
	"core-RPC/server"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func startServer() {
	lis, _ := net.Listen("tcp", ":8849")
	server.Accept(lis)
}
func main() {
	go startServer()
	conn, _ := net.Dial("tcp", ":8849")
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)

	_ = json.NewEncoder(conn).Encode(server.DefaultOption)
	c := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Jiyeon's Method",
			Seq:           uint64(i),
		}
		_ = c.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = c.ReadHeader(h)
		var reply string
		_ = c.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
