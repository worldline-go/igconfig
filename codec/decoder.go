package codec

import (
	"fmt"
	"io"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/reformat.git"
)

// Decoder is a interface that will decode reader `r` into `to`.
//
// `to` will already be pointer to struct, so no need to further prepare it for decoding.
type Decoder interface {
	Decode(r io.Reader, to interface{}) error
}

// MapDecoder implements the reformat package,
// it exposes functionality to convert an arbitrary map[string]interface{}
// into a native Go structure with given tag name.
func MapDecoder(input, output interface{}, tag string) error {
	cnf := &reformat.DecoderConfig{
		DecodeHook:       nil,
		ErrorUnused:      false,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		Metadata:         nil,
		Result:           output,
		TagName:          tag,
	}

	decoder, err := reformat.NewDecoder(cnf)
	if err != nil {
		return fmt.Errorf("could not create new decoder: %w", err)
	}

	return decoder.Decode(input)
}
