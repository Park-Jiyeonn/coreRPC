package server

import (
	"core-RPC/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 114514

type Option struct {
	MagicNumber int
	CodecType   string
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

var DeafaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: Accept error", err)
			return
		}
		// 利用 conn 去处理消息
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) {
	DeafaultServer.Accept(lis)
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: option error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	// 构造好 Gob了，是实现了 codec 接口的 gob，可以去接收和发送消息了
	server.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(c codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	//您说得对，这段代码中的 for 循环会一直运行，因此主goroutine不会执行Wait方法。
	//实际上，在这个RPC服务器中,
	//编解码器的关闭是由客户端发起的，因此服务器需要一直等待客户端关闭连接，才能关闭编解码器。
	//
	//因此，在这种情况下，使用WaitGroup的主要作用是确保所有请求都已经处理完成，
	//而不是等待所有goroutine退出。
	//当所有请求处理完成后，主goroutine会一直等待客户端关闭连接，直到客户端关闭连接后，
	//主goroutine才会执行编解码器的关闭操作。这样可以确保所有请求都得到了正确的处理，并且服务器在关闭之前不会丢失任何请求。

	for {
		req, err := server.readRequest(c)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(c, req.h, invalidRequest, sending)
		}
		wg.Add(1)
		go server.handleRequest(c, req, sending, wg)
	}
	wg.Wait()
	_ = c.Close()
}

type request struct {
	h           *codec.Header
	argv, reply reflect.Value
}

func (server *Server) ReadRequestHeader(c codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := c.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read head error", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(c codec.Codec) (*request, error) {
	h, err := server.ReadRequestHeader(c)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = c.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err: ", err)
	}
	return req, nil
}

func (server *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(h, body); err != nil {
		log.Println("rpc server: write response error: ", err)
	}
}

func (server *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.reply = reflect.ValueOf(fmt.Sprintf("core-rpc %d", req.h.Seq))
	server.sendResponse(c, req.h, req.reply.Interface(), sending)
}
