package aioli

import (
	"io"
)

// Decoder interface to be used with I/O manager Listen method.
type Decoder interface {
	Decode(v interface{}) error
}

// NewDecoder returns the default implemenation JSONDecoder
func NewDecoder(r io.Reader) Decoder {
	return NewJSONDecoder(r)
}
