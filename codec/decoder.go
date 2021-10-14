package codec

import (
	"io"
)

// Decoder is a interface that will decode reader `r` into `to`.
//
// `to` will already be pointer to struct, so no need to further prepare it for decoding.
type Decoder interface {
	Decode(r io.Reader, to interface{}) error
}
