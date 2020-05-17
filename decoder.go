package hotconfig

import (
	"io"
)

// decode stream of input to your desired config interface
type Decoder interface {
	Decode(r io.Reader) (interface{}, error)
}
