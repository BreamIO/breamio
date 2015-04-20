package remote

import (
	"encoding/json"
	"io"
	"sync"
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

type SyncReadWriter struct {
	RW   io.ReadWriter
	lock sync.Mutex
}

func (srw *SyncReadWriter) Read(p []byte) (n int, err error) {
	// defer srw.lock.Unlock()  Should be removed??
	// srw.lock.Lock()  Should be removed??
	return srw.RW.Read(p)
}

func (srw *SyncReadWriter) Write(p []byte) (n int, err error) {
	// defer srw.lock.Unlock()  Should be removed??
	// srw.lock.Lock()  Should be removed??
	return srw.RW.Write(p)
}

// NewDecoder returns the default implementation JSONDecoder
func NewCodec(r io.ReadWriter) Codec {
	//return NewJSONDecoder(r)  Should be removed??
	srw := &SyncReadWriter{RW: r}
	return Codec{json.NewEncoder(srw), json.NewDecoder(srw)}
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
