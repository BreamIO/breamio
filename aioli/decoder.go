package aioli

import (
	"encoding/json"
	"io"
)

type EncodeDecoder interface {
	Encoder
	Decoder
}

type Codec struct {
	Encoder
	Decoder
}

// NewDecoder returns the default implemenation JSONDecoder
func NewCodec(r io.ReadWriter) Codec {
	//return NewJSONDecoder(r)
	return Codec{json.NewEncoder(r), json.NewDecoder(r)}
}

// Decoder interface to be used with I/O manager Listen method.
type Decoder interface {
	Decode(v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) error
}
