package internal

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type smallInnerStruct struct {
	C   int           `cfg:"c" env:"inner"`
	Dur time.Duration `cfg:"dur"`
}

type structWithEverything struct {
	A string            `cfg:"a"`
	B smallInnerStruct  `cfg:"b" env:"struct"`
	C *smallInnerStruct `cfg:"c"`
	D int               `cfg:"d"`
}

func TestEnvFieldName(t *testing.T) {
	tests := []struct {
		name   string
		outer  string
		tag    string
		result string
	}{
		{
			name:   "simple field",
			tag:    `cfg:"field"`,
			result: "FIELD",
		},
		{
			name:   "with comma",
			tag:    `cfg:"long,l"`,
			result: "LONG",
		},
		{
			name:   "with base and long",
			outer:  "test",
			tag:    `cfg:"long,l"`,
			result: "TEST_LONG",
		},
	}
	envFieldName := FieldNameWithSeparator("env", "_", strings.ToUpper)
	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			result := envFieldName(test.outer, reflect.StructField{Tag: reflect.StructTag(test.tag)})
			assert.Equal(t, test.result, result)
		})
	}
}

func TestStructIterator_Iterate(t *testing.T) {
	tests := []struct {
		Name     string
		Iterator StructIterator
		Result   interface{}
		Error    string
	}{
		{
			Name: "small normal env",
			Iterator: StructIterator{
				Value:         &structWithEverything{},
				FieldNameFunc: FieldNameWithSeparator("env", "_", strings.ToUpper),
				IteratorFunc: mapIterator(map[string]string{
					"A": "1",
				}),
			},
			Result: &structWithEverything{A: "1", C: &smallInnerStruct{}},
		},
		{
			Name: "input is not pointer",
			Iterator: StructIterator{
				Value: structWithEverything{},
			},
			Error: ErrInputIsNotPointerOrStruct.Error(),
		},
		{
			Name: "input is not struct",
			Iterator: StructIterator{
				Value: new(int),
			},
			Error: ErrInputIsNotPointerOrStruct.Error(),
		},
		{
			Name: "env with inner struct",
			Iterator: StructIterator{
				Value:         &structWithEverything{},
				FieldNameFunc: FieldNameWithSeparator("env", "_", strings.ToUpper),
				IteratorFunc: mapIterator(map[string]string{
					"A":            "1",
					"STRUCT_INNER": "5",
					"C_DUR":        "3s",
				}),
			},
			Result: &structWithEverything{A: "1", B: smallInnerStruct{C: 5}, C: &smallInnerStruct{Dur: 3 * time.Second}},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			err := test.Iterator.Iterate()

			if test.Error == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.Result, test.Iterator.Value)
			} else {
				// Value is unusable in this case, no reason to check it.
				assert.EqualError(t, err, test.Error)
			}

		})
	}
}

func TestStructIterator_FieldNames(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		fieldFunc FieldNameFunc
		result    []string
	}{
		{
			name: "no tags",
			input: &struct {
				Inner struct {
					FieldInner bool
					unexported []string
				}
				Field1 string
			}{},
			fieldFunc: FieldNameWithSeparator("", "_", strings.ToUpper),
			result:    []string{"INNER", "INNER_FIELDINNER", "FIELD1"},
		},
		{
			name: "some tags",
			input: &struct {
				Inner struct {
					FieldInner bool `env:"i"`
					unexported []string
				} `cfg:"rename"`
				Field1 string `cmd:"flag"`
			}{},
			fieldFunc: FieldNameWithSeparator("env", "_", strings.ToUpper),
			result:    []string{"RENAME", "RENAME_I", "FIELD1"},
		},
		{
			name: "skip struct",
			input: &struct {
				Inner struct {
					FieldInner bool `env:"i"`
					unexported []string
				} `cfg:"-"`
				Field1 string `cmd:"flag"`
			}{},
			fieldFunc: FieldNameWithSeparator("env", "_", strings.ToUpper),
			result:    []string{"-", "FIELD1"},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			fields := make([]string, 0, len(test.result))

			iter := StructIterator{
				FieldNameFunc: fieldCapture(test.fieldFunc, &fields),
				IteratorFunc:  func(fieldName string, field reflect.Value) error { return nil },
				Value:         test.input,
			}

			require.NoError(t, iter.Iterate())

			require.Equal(t, test.result, fields)
		})
	}

}

func fieldCapture(nameFunc FieldNameFunc, namesSlice *[]string) FieldNameFunc {
	return func(outerName string, currentField reflect.StructField) string {
		name := nameFunc(outerName, currentField)

		*namesSlice = append(*namesSlice, name)

		return name
	}
}

func mapIterator(fieldMap map[string]string) IteratorFunc {
	return func(fieldName string, field reflect.Value) error {
		val, ok := fieldMap[fieldName]
		if !ok {
			return nil
		}

		return SetReflectValueString(fieldName, val, field)
	}
}
