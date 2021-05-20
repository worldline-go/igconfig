package igconfig

import (
	"errors"
	"strings"
	"testing"
	"time"
	"fmt"

	"github.com/stretchr/testify/assert"

	"github.com/rs/zerolog"
)

type small struct {
	Int1         int
	String       string
	unexported   string
	timeField    time.Time
	Zerolog      *zerologMarshaler
	NonPrintable string        `loggable:"false"`
	NonPrintableSecret string  `secret:"i_like_turtles"`
}

type Secret struct  {
	NonPrintableSecret string  `secret:"i_like_turtles"`
	PrintableSecret    string  `secret:"i_like_turtles" loggable:"true"`
}
type withTimeFields struct {
	Int           int
	Bool          bool
	Float64       float64
	TimeField     time.Time
	DurationField time.Duration
}

type innerStruct struct {
	First  small
	Second *small
}

type withMarshaler struct {
	Marshaled TestError
}

type zerologMarshaler struct {
	Int int
}

func (z *zerologMarshaler) MarshalZerologObject(e *zerolog.Event) {
	e.Int("super_cool_field", z.Int)
}

type TestError string

func TestPrinter_MarshalZerologObject(t *testing.T) {
	tests := []struct {
		Name            string
		LoggableTagName string
		Value           interface{}
		Result          string
	}{
		{
			Name:   "no value",
			Result: "{}",
		},
		{
			Name:   "small",
			Value:  small{Int1: 3, String: "str", unexported: "test"},
			Result: `{"int1":3,"string":"str","zerolog":null}`,
		},
		{
			Name:   "dont print secrets",
			Value:  small{Int1: 3, String: "str", unexported: "test", NonPrintableSecret: "dont_log_me"},
			Result: `{"int1":3,"string":"str","zerolog":null}`,
		},
		{
			Name:   "with time",
			Value:  withTimeFields{DurationField: 66 * time.Second},
			Result: `{"int":0,"bool":false,"float64":0,"timefield":"0001-01-01T00:00:00Z","durationfield":"1m6s"}`,
		},
		{
			Name:   "nil inner struct",
			Value:  innerStruct{First: small{String: "inner_string"}},
			Result: `{"first":{"int1":0,"string":"inner_string","zerolog":null},"second":null}`,
		},
		{
			Name:   "inner struct",
			Value:  innerStruct{First: small{String: "inner_string"}, Second: &small{String: "second"}},
			Result: `{"first":{"int1":0,"string":"inner_string","zerolog":null},"second":{"int1":0,"string":"second","zerolog":null}}`,
		},
		{
			Name:   "pointer struct",
			Value:  &small{String: "inner_string"},
			Result: `{"int1":0,"string":"inner_string","zerolog":null}`,
		},
		{
			Name:   "nil pointer struct",
			Value:  (*small)(nil),
			Result: `{}`,
		},
		{
			Name:   "scalar type",
			Value:  0,
			Result: `{}`,
		},
		{
			Name:   "with marshaler error",
			Value:  withMarshaler{Marshaled: "test error"},
			Result: `{"error_marshaled":"test error"}`,
		},
		{
			Name:   "with marshaler no error",
			Value:  withMarshaler{},
			Result: `{"marshaled":"valid"}`,
		},
		{
			Name:   "zerolog Object marshaler",
			Value:  &zerologMarshaler{Int: 6},
			Result: `{"super_cool_field":6}`,
		},
		{
			Name:   "zerolog do not implement Object marshaler",
			Value:  zerologMarshaler{Int: 6},
			Result: `{"int":6}`,
		},
		{
			Name:   "small with zerolog marshable value",
			Value:  small{Int1: 3, Zerolog: &zerologMarshaler{Int: 33}, unexported: "test"},
			Result: `{"int1":3,"string":"","zerolog":{"super_cool_field":33}}`,
		},
	}

	var b strings.Builder
	logger := zerolog.New(&b)

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			logger.Log().EmbedObject(Printer{LoggableTag: test.LoggableTagName, Value: test.Value}).Send()

			assert.Equal(t, test.Result, strings.TrimSpace(b.String()),fmt.Sprintf("test name %s, value %s",test.Name,strings.TrimSpace(b.String())))

			b.Reset()
		})
	}
}

func (e TestError) MarshalText() ([]byte, error) {
	if e != "" {
		return nil, errors.New(string(e))
	}

	return []byte("valid"), nil
}

func Test_secret_logging (t *testing.T) {
	tests := []struct {
		Name            string
		LoggableTagName string
		Value           interface{}
		Result          string
	}{
		{
			Name:   "to log or not to log",
			Value:  Secret{PrintableSecret: "log_me", NonPrintableSecret: "dont_log_me"},
			Result: `{"printablesecret":"log_me"}`,
		},
	}
	var b strings.Builder
	logger := zerolog.New(&b)

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			logger.Log().EmbedObject(Printer{LoggableTag: test.LoggableTagName, Value: test.Value}).Send()

			assert.Equal(t, test.Result, strings.TrimSpace(b.String()),fmt.Sprintf("test name %s, value %s",test.Name,strings.TrimSpace(b.String())))

			b.Reset()
		})
	}
}

