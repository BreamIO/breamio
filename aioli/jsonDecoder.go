package aioli

import (
	"encoding/json"
	"encoding/gob"
	"io"
)

// JSONDecoder implements Decoder interface and uses gob and json for decoding // data
type JSONDecoder struct {
	reader io.Reader
	gobdec *gob.Decoder
}

// NewJSONDecoder returns a pointer to a JSONDecoder
func NewJSONDecoder(r io.Reader) *JSONDecoder {
	return &JSONDecoder{
	reader: r,
	gobdec: gob.NewDecoder(r)}
}

// Decode tries to decode incomming data into the underlying value of v
//
// Parameter v must be a pointer value.
func (jd *JSONDecoder) Decode(v interface{}) error {
	var data []byte
	err := jd.gobdec.Decode(&data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	return nil
}
