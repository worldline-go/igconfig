package codec

import (
	"reflect"
	"time"

	"github.com/worldline-go/struct2"
	"github.com/xhit/go-str2duration/v2"
)

// BackupTagName is the tag name used if tagname not found in the struct tags.
var BackupTagName = "cfg"

var (
	// WeaklyIgnoreSeperator is used to ignore the seperator in the tag name.
	WeaklyIgnoreSeperator = true
	// WeaklyDashUnderscore is used to convert dash to underscore in the tag name.
	// If WeaklyIgnoreSeperator is set to true then this will be ignored.
	WeaklyDashUnderscore = false

	// HooksDecode functions to convert value before decoding.
	HooksDecode = []struct2.HookDecodeFunc{
		// Convert string to time.Duration
		func(t1, t2 reflect.Type, data interface{}) (interface{}, error) {
			if t2 != reflect.TypeOf(time.Duration(0)) {
				return data, nil
			}

			switch t1.Kind() {
			case reflect.String:
				return str2duration.ParseDuration(data.(string))
			case reflect.Int:
				return time.Duration(data.(int)), nil
			case reflect.Int64:
				return time.Duration(data.(int64)), nil
			case reflect.Float64:
				return time.Duration(data.(float64)), nil
			default:
				return data, nil
			}
		},
	}
)

// MapDecoder implements the reformat package,
// it exposes functionality to convert an arbitrary map[string]interface{}
// into a native Go structure with given tag name.
func MapDecoder(input, output interface{}, tag string) error {
	decoder := struct2.Decoder{
		TagName:               tag,
		BackupTagName:         BackupTagName,
		HooksDecode:           HooksDecode,
		WeaklyTypedInput:      true,
		WeaklyIgnoreSeperator: WeaklyIgnoreSeperator,
		WeaklyDashUnderscore:  WeaklyDashUnderscore,
	}

	return decoder.Decode(input, output)
}
