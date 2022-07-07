package codec

import (
	"io"

	"gopkg.in/yaml.v3"
)

// YAML is a yaml decoder.
type YAML struct {
	Strict bool
}

// Decode is a decoder function for yaml.
func (c YAML) Decode(r io.Reader, to interface{}) error {
	decoder := yaml.NewDecoder(r)

	if c.Strict {
		decoder.KnownFields(true)
	}

	return decoder.Decode(to)
}

var _ Decoder = YAML{}
