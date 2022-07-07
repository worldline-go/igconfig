package codec

import (
	"io"

	"github.com/BurntSushi/toml"
)

// TOML is a toml decoder.
type TOML struct{}

// Decode is a decoder function for yaml.
func (c TOML) Decode(r io.Reader, to interface{}) error {
	decoder := toml.NewDecoder(r)

	_, err := decoder.Decode(to)

	return err
}

var _ Decoder = YAML{}
