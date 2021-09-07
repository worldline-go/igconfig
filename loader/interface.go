package loader

import (
	"context"
	"io"
	"time"

	"gopkg.in/yaml.v2"
)

// DefaultDecoder specifies which decoder will be used in specific cases.
// For example in Consul and Reader.
var DefaultDecoder Decoder = StrictYamlDecoder

// Decoder is a function that will decode reader `r` into `to`.
//
// `to` will already be pointer to struct, so no need to further prepare it for decoding.
type Decoder func(r io.Reader, to interface{}) error

type DynamicRunner func(value []byte) error

type DynamicConfig struct {
	AppName string
	// FieldName that should be looked up.
	//
	// This is not exactly field name, rather how it will be named in Consul.
	// For example in Consul path /finops/adm0001s/loglevel FieldName MUST be set to "loglevel".
	FieldName       string
	RefreshInterval time.Duration
	// Runner will be executed when new, different value will be received.
	//
	// Returning error will stop dynamic updates, so it should be restarted manually.
	Runner DynamicRunner
}

type Loader interface {
	// Load will load all available data to at 'to' value.
	//
	// Even if particular loader type must implement ReflectLoader -
	// this interface still must be implemented as a proxy.
	Load(appName string, to interface{}) error
}

type DynamicValuer interface {
	// DynamicValue polls config field value once in dynamicConfig.RefreshInterval.
	// If after poll value was changed - dynamicConfig.Runner function will be called with new value.
	//
	// If requested service supports better(native) listening for changes - it will be implemented instead.
	//
	// Error handling should be done in runner function.
	DynamicValue(ctx context.Context, dynamicConfig DynamicConfig) error
}

// StrictYamlDecoder is a decoder function.
func StrictYamlDecoder(r io.Reader, to interface{}) error {
	decoder := yaml.NewDecoder(r)
	decoder.SetStrict(true)

	return decoder.Decode(to)
}
