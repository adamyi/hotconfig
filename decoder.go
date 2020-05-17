package hotconfig

import (
	"io"
)

// decode stream of input to your desired config interface
type Decoder interface {
	Decode(r io.Reader) (interface{}, error)
}

// a simple decoder that calls a function
type DecoderFunc func(r io.Reader) (interface{}, error)

func (f DecoderFunc) Decode(r io.Reader) (interface{}, error) {
	return f(r)
}
