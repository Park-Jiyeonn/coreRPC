package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
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

func main() {
	addr := make(chan string)
	go client(addr)
	server(addr)
	//time.Sleep(3 * time.Second)
}
