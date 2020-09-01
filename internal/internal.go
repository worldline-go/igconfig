package internal

import (
	"errors"
	"reflect"
	"strings"
)

const DefaultConfigTag = "cfg"

// TagValue returns splitted by comma value of tag for specified key.
// If no such key in tag is present - DefaultConfigTag will be used to get value.
//
// Note:
//	TagValue(field, "key") == TagValueByKeys(field, "key", DefaultConfigTag)
//
// See TagValueByKeys for usage examples.
func TagValue(field reflect.StructField, key string) []string {
	return TagValueByKeys(field, key, DefaultConfigTag)
}

// TagValueByKeys returns comma-separated list of values based on keys.
//
// Keys are sorted, which means that if first key has valid value - that value will be returned and
// no further keys will be checked.
//
// This function is useful to get default tag value if no requested tags have any values:
//	tagVal := TagValueByKeys(field, "highestPriority", "lowerPriority", "lowestPriority", "defaultTagKey")
//
// It is valid to call this function with zero or one key.
//
// If requested tag has a value of '-'(minus sign) - no further keys will be checked and nil will be returned.
//
// If no requested keys have valid values - single value of lower-cased field name will be returned.
func TagValueByKeys(field reflect.StructField, keys ...string) []string {
	var tagValue string
	for _, tagName := range keys {
		tagValue = field.Tag.Get(tagName)
		if tagValue == "-" {
			return nil
		}

		if tagValue != "" {
			break
		}
	}

	if tagValue == "" {
		return []string{strings.ToLower(field.Name)}
	}

	return strings.Split(tagValue, ",")
}

func GetReflectElem(val interface{}) (reflect.Value, error) {
	refVal := reflect.ValueOf(val)

	if refVal.Kind() != reflect.Ptr {
		return reflect.Value{}, errors.New("value is not pointer")
	}

	return refVal.Elem(), nil
}
