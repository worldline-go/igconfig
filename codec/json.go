package codec

import (
	"encoding/json"
	"io"
)

// JSON is a json decoder.
type JSON struct {
	Strict bool
}

// Decode is a decoder function for json.
func (c JSON) Decode(r io.Reader, to interface{}) error {
	decoder := json.NewDecoder(r)

	if c.Strict {
		decoder.DisallowUnknownFields()
	}

	return decoder.Decode(to)
}

var _ Decoder = JSON{}
