package igconfig

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = Printer{}

const LoggableTagName = "loggable"

type NameGetter func(t reflect.StructField) string

// Printer is an implementation of zerolog.LogObjectMarshaler for marshaling struct to zerolog event.
//
// Primary use case is to add configuration struct to the log.
//
// Usage example:
//	conf := config.AppConfig{} // set up config value somehow
//	// log is zerolog/log package
//	log.Info().
//		Object("config", Printer{Value: conf}).
//		Msg("loaded config")
type Printer struct {
	// LoggableTag will be used to get tag value and check if field can be logged or not.
	LoggableTag string
	// NameGetter will be called for each field to get name of it.
	NameGetter
	// Value is actual struct that should be printed.
	// It is possible for the value to be pointer to a struct
	Value interface{}
}

// MarshalZerologObject marshals config to zerolog event.
//
// Value should not be specifically config, but could be any struct.
func (p Printer) MarshalZerologObject(ev *zerolog.Event) {
	e, ok := p.validateInput()
	if !ok {
		return
	}

	t := e.Type()

	if p.LoggableTag == "" {
		p.LoggableTag = LoggableTagName
	}

	if p.NameGetter == nil {
		p.NameGetter = DefaultNameGetter
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		b, ok := f.Tag.Lookup(p.LoggableTag)

		// Currently only 'true' is supported as value for LoggableTag
		if ok && b != "true" {
			continue
		}

		name, elem := p.NameGetter(f), e.FieldByName(f.Name)

		// Do not log unexported fields.
		if !elem.CanInterface() {
			continue
		}

		if ok := p.printInterfacePrinter(ev, name, elem); ok {
			continue
		}

		ev = p.logTyped(ev, name, elem)
	}
}

// logTyped will add elements to zerolog event based on type of value.
func (p Printer) logTyped(ev *zerolog.Event, name string, elem reflect.Value) *zerolog.Event {
	switch elem.Kind() {
	case reflect.String:
		return ev.Str(name, elem.String())
	case reflect.Int64, reflect.Int:
		return ev.Int64(name, elem.Int())
	case reflect.Float64:
		return ev.Float64(name, elem.Float())
	case reflect.Bool:
		return ev.Bool(name, elem.Bool())
	default:
		// If element is a non-nil pointer to a struct - convert it to a struct and print it.
		if elem.Kind() == reflect.Ptr && !elem.IsNil() && elem.Elem().Kind() == reflect.Struct {
			elem = elem.Elem()
		}

		if elem.Kind() == reflect.Struct {
			return ev.Object(name, Printer{
				LoggableTag: p.LoggableTag,
				NameGetter:  p.NameGetter,
				Value:       elem.Interface(),
			})
		}

		return ev.Interface(name, elem.Interface())
	}
}

// printInterfacePrinter checks if value implements well-known print interfaces.
//
// Currently it checks for encoding.TextMarshaler and fmt.Stringer
//
// If some interface method will return error - field name will have form of 'error_<field_name>'
// (where <field_name> will be returned from NameGetter) and error will be the value.
func (p Printer) printInterfacePrinter(ev *zerolog.Event, name string, elem reflect.Value) bool {
	interfaceValue := elem.Interface()

	if textMarshaler, ok := interfaceValue.(encoding.TextMarshaler); ok {
		txt, err := textMarshaler.MarshalText()
		if err != nil {
			ev = ev.Str("error_"+name, ""+err.Error())
			return true
		}

		ev = ev.Str(name, string(txt))
		return true
	}

	if stringer, ok := interfaceValue.(fmt.Stringer); ok {
		ev = ev.Stringer(name, stringer)
		return true
	}

	return false
}

// validateInput checks that Printer input value is valid.
//
// Valid means that it is a struct or non-nil pointer to a struct
func (p Printer) validateInput() (reflect.Value, bool) {
	if p.Value == nil {
		return reflect.Value{}, false
	}

	e := reflect.ValueOf(p.Value)

	if e.Kind() == reflect.Ptr && e.IsNil() {
		return reflect.Value{}, false
	}

	e = reflect.Indirect(e)
	t := e.Type()

	if t.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	return e, true
}

// DefaultNameGetter returns lowercase field name as json field name.
func DefaultNameGetter(t reflect.StructField) string {
	return strings.ToLower(t.Name)
}
