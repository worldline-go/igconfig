package igconfig

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = Printer{}

var (
	// LogTagOptionNames is a tag name for loggable boolean check.
	LogTagOptionNames = []string{"loggable", "log"}
	// SecretTagName is a tag name for secret loaders to prevent print it.
	SecretTagName = "secret"
)

//nolint:golint
type NameGetter func(t reflect.StructField) string

// Printer is an implementation of zerolog.LogObjectMarshaler for marshaling struct to zerolog event.
//
// Primary use case is to add configuration struct to the log.
//
// Usage example:
//
//	conf := config.AppConfig{} // set up config value somehow
//	// log is zerolog/log package
//	log.Info().
//		Object("config", Printer{Value: conf}).
//		Msg("loaded config")
type Printer struct {
	// NameGetter will be called for each field to get name of it.
	NameGetter
	// Value is actual struct that should be printed.
	// It is possible for the value to be pointer to a struct.
	//
	// Value can implement some zerolog interfaces, but it has to be of proper type:
	// if interface is defined on pointer receiver - passed value should also be pointer.
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

	if marshaler, ok := p.Value.(zerolog.LogObjectMarshaler); ok {
		ev.EmbedObject(marshaler)

		return
	}

	t := e.Type()

	if p.NameGetter == nil {
		p.NameGetter = DefaultNameGetter
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		var logValue string
		var ok bool
		for _, tag := range LogTagOptionNames {
			logValue, ok = f.Tag.Lookup(tag)
			if ok {
				break
			}
		}

		loggable, _ := strconv.ParseBool(logValue)

		if ok && !loggable {
			continue
		}

		secretValues, isSecret := f.Tag.Lookup(SecretTagName)

		fmt.Println(secretValues, isSecret, loggable)

		// If the tag could potentially be a secret you need to
		// explicitly state that you want to log it
		// else the default is not log it.
		if isSecret && !loggable {
			skipThisField := false
			for _, logValue := range LogTagOptionNames {
				if !isInTagOption(secretValues, logValue) {
					skipThisField = true
				}
			}

			if skipThisField {
				continue
			}
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
				NameGetter: p.NameGetter,
				Value:      elem.Interface(),
			})
		}

		return ev.Interface(name, elem.Interface())
	}
}

// printInterfacePrinter checks if value implements well-known print interfaces.
//
// Currently it checks for encoding.TextMarshaler and fmt.Stringer.
//
// If some interface method will return error - field name will have form of 'error_<field_name>'
// (where <field_name> will be returned from NameGetter) and error will be the value.
func (p Printer) printInterfacePrinter(ev *zerolog.Event, name string, elem reflect.Value) bool {
	// Print null if pointer is nil.
	// There is great chance to get panic with interface that don't handle nil receiver.
	if !elem.IsValid() || (elem.Kind() == reflect.Ptr && elem.IsNil()) {
		// Maybe this will be changed in the future, but currently let's be verbose about nil pointers.
		ev.Interface(name, nil)

		return true
	}

	interfaceValue := elem.Interface()

	switch v := interfaceValue.(type) {
	case zerolog.LogObjectMarshaler:
		ev.Object(name, v)
	case zerolog.LogArrayMarshaler:
		ev.Array(name, v)
	case json.Marshaler:
		jsn, err := v.MarshalJSON()
		if err != nil {
			ev.AnErr("error_"+name, err)

			return true
		}

		ev.RawJSON(name, jsn)
	case encoding.TextMarshaler:
		txt, err := v.MarshalText()
		if err != nil {
			ev.AnErr("error_"+name, err)

			return true
		}

		ev.Bytes(name, txt)
	case fmt.Stringer:
		ev.Stringer(name, v)
	default:
		return false
	}

	return true
}

// validateInput checks that Printer input value is valid.
//
// Valid means that it is a struct or non-nil pointer to a struct.
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

func isInTagOption(tagValues, check string) bool {
	printSecret := false
	splitSecret := strings.Split(tagValues, ",")

	if len(splitSecret) > 1 {
		for i := 1; i < len(splitSecret); i++ {
			if strings.TrimSpace(splitSecret[i]) == check {
				printSecret = true

				break
			}
		}
	}

	return printSecret
}
