package aioli

import (
	"encoding/json"
	"io"
)

// ReadWriter is the interface that groups the Encode and Decode methods.
type EncodeDecoder interface {
	Encoder
	Decoder
}

// A practical implementation of EncodeDecoder.
//
// Embeds a Encoder and a Decoder.
// This allows preexisting encodings where the methods are split on different types to be used as a EncoderDecoder.
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

// Encoder wraps the Encode method.
//
// The signature is chosen to allow encoding/json to be used without wrapping.
// And because it is practical and makes sense in the intended use case.
type Encoder interface {
	Encode(v interface{}) error
}
