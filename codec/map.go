package codec

import (
	"fmt"

	"github.com/worldline-go/reformat"
)

// BackupTagName is the tag name used if tagname not found in the struct tags.
var BackupTagName = "cfg"

// MapDecoder implements the reformat package,
// it exposes functionality to convert an arbitrary map[string]interface{}
// into a native Go structure with given tag name.
func MapDecoder(input, output interface{}, tag string) error {
	cnf := &reformat.DecoderConfig{
		DecodeHook:           nil,
		ErrorUnused:          false,
		ZeroFields:           false,
		WeaklyTypedInput:     true,
		Metadata:             nil,
		Result:               output,
		TagName:              tag,
		BackupTagName:        BackupTagName,
		WeaklyDashUnderscore: true,
	}

	decoder, err := reformat.NewDecoder(cnf)
	if err != nil {
		return fmt.Errorf("could not create new decoder: %w", err)
	}

	return decoder.Decode(input)
}
