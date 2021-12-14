package codec

import (
	"fmt"
	"io"
)

// LoadReaderWithDecoder will decode input in `r` into `to` by using `decoder`.
func LoadReaderWithDecoder(r io.Reader, to interface{}, decoder Decoder, tag string) error {
	mapping := map[string]interface{}{}
	if err := decoder.Decode(r, &mapping); err != nil {
		return fmt.Errorf("LoadReaderWithDecoder: decoder.Decode error: %w", err)
	}

	if err := MapDecoder(&mapping, to, tag); err != nil {
		return fmt.Errorf("LoadReaderWithDecoder codec.MapDecoder error: %w", err)
	}

	return nil
}
