package codec

import (
	"io"

	"gopkg.in/yaml.v2"
)

type YAML struct {
	Strict bool
}

// Decode is a decoder function for yaml.
func (c YAML) Decode(r io.Reader, to interface{}) error {
	decoder := yaml.NewDecoder(r)

	if c.Strict {
		decoder.SetStrict(true)
	}

	return decoder.Decode(to)
}

var _ Decoder = YAML{}
