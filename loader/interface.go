package loader

import (
	"context"
	"errors"
)

// ErrNoClient is returned when no client is found, for Vault and Consul.
var ErrNoClient = errors.New("no client available")

// Loader is a interface for all loaders.
type Loader interface {
	// Load will load all available data to at 'to' value.
	//
	// Even if particular loader type must implement ReflectLoader -
	// this interface still must be implemented as a proxy.
	Load(appName string, to interface{}) error

	// LoadWithContext same as Load but using predefined ctx in load process.
	// This is usable for logging.
	LoadWithContext(ctx context.Context, appName string, to interface{}) error
}

// DynamicValuer interface is used to get dynamic value from loader.
type DynamicValuer interface {
	// DynamicValue polls config field value once in dynamicConfig.RefreshInterval.
	// If after poll value was changed - dynamicConfig.Runner function will be called with new value.
	//
	// If requested service supports better(native) listening for changes - it will be implemented instead.
	//
	// Error handling should be done in runner function.
	DynamicValue(context.Context, string) (<-chan []byte, error)
}
