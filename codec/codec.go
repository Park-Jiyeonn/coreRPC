package codec

import "io"

type Header struct {
	ServiceMethod string // 请求的方法名
	Seq           uint64 // 请求 ID
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

const (
	GobType  string = "application/gob"
	JsonType string = "application/json"
)

var NewCodecFuncMap map[string]NewCodecFunc

// init 的作用就是，我们需要 Codec 的实例的时候，通过这个 Type 可以从 Map 中得到构造函数
func init() {
	NewCodecFuncMap = make(map[string]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
