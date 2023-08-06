package codec

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type JSONCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *json.Decoder
	enc  *json.Encoder
}

func NewJSONCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &JSONCodec{
		conn: conn,
		buf:  buf,
		dec:  json.NewDecoder(conn),
		enc:  json.NewEncoder(buf),
	}
}

// 开始实现接口定义的方法。

func (c *JSONCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *JSONCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *JSONCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err = c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header", err)
		return err
	}
	if err = c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body", err)
		return err
	}
	return nil
}

func (c *JSONCodec) Close() error {
	return c.conn.Close()
}
