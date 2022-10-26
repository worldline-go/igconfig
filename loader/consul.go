package loader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/consul/api"
	"github.com/worldline-go/igconfig/codec"
	"github.com/worldline-go/igconfig/internal"
)

// ConsulConfigPathPrefixEnv holds the name of the env variable that is used to set custom path prefix.
const ConsulConfigPathPrefixEnv = "CONSUL_CONFIG_PATH_PREFIX"

// ConsulTag is a tag used to identify field name.
var ConsulTag = "cfg"

// ConsulConfigPathPrefix stores the default base path for secrets.
var ConsulConfigPathPrefix = "finops"

var _ Loader = Consul{}

var _ DynamicValuer = Consul{}

// LiveServiceFetcher is a signature of the function that will fetch only live instances of the service.
//
// If no services found - (nil, nil) will be returned.
type LiveServiceFetcher func(ctx context.Context, name string, tags []string) ([]*api.ServiceEntry, error)

// Consul is an instance of configuration loader from Consul.
//
// Example usage:
//
//	var config Config // some Config struct
//
//	cl, err := api.NewClient(&api.Config{Address: "http://consul:8500"})
//	if err != nil { ... }
//
//	consulLoader := Consul{Client: cl}
//	err = consulLoader.Load("adm0001s", &config)
//	if err != nil { ... }
//
//	// config is now populated from Consul.
type Consul struct {
	Client *api.Client
	// Decoder specifies function that will decode the response from Consul.
	// By default it is YAML parser + Map decoder.
	//
	// Please prefer YAML to JSON or anything else if there is no strict requirement for it.
	//
	// Note: this function is not used in Watcher.
	Decoder codec.Decoder
	// Plan for dynamic changes
	Plan Planer
}

// LoadWithContext retrieves data from Consul and decode response into 'to' struct.
func (l Consul) LoadWithContext(ctx context.Context, appName string, to interface{}) error {
	if err := l.EnsureClient(); err != nil {
		return err
	}

	queryOptions := api.QueryOptions{}
	data, _, err := l.Client.KV().Get(
		path.Join(internal.GetEnvWithFallback(ConsulConfigPathPrefixEnv, ConsulConfigPathPrefix), appName),
		queryOptions.WithContext(ctx),
	)
	// If no data or err is returned - return early.
	if data == nil || err != nil {
		return err
	}

	if l.Decoder == nil {
		l.Decoder = codec.YAML{}
	}

	if err := codec.LoadReaderWithDecoder(bytes.NewReader(data.Value), to, l.Decoder, ConsulTag); err != nil {
		return fmt.Errorf("Consul.LoadWithContext error: %w", err)
	}

	return nil
}

// Load is just same as LoadWithContext without context.
func (l Consul) Load(appName string, to interface{}) error {
	return l.LoadWithContext(context.Background(), appName, to)
}

// EnsureClient creates and sets a Consul client if needed.
func (l *Consul) EnsureClient() error {
	if l.Client == nil {
		var err error

		l.Client, err = NewConsulFromEnv()
		if err != nil {
			return err
		}
	}

	if l.Client == nil {
		return ErrNoClient
	}

	return nil
}

// SearchLiveServices is a wrapper for c.Client.Health().ServiceMultipleTags(name, tags, true, (&api.QueryOptions{}).WithContext(ctx))
//
// This provides a bit nicer interface on fetching services
// and gives ability to have LiveServiceFetcher as an argument or a field instead of actual implementation.
func (l *Consul) SearchLiveServices(ctx context.Context, name string, tags []string) ([]*api.ServiceEntry, error) {
	if err := l.EnsureClient(); err != nil {
		return nil, err
	}

	services, _, err := l.Client.Health().ServiceMultipleTags(name, tags, true, (&api.QueryOptions{}).WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("fetch service instances: %w", err)
	}

	return services, nil
}

// NewConsul creates a client from a client.
func NewConsul(addr string) (*api.Client, error) {
	return NewConsulWithConfig(&api.Config{Address: addr})
}

// NewConsulFromEnv creates a client from environmental variables.
//
// This function uses api.DefatulConfig(), which means that variables should be named as Consul expects them.
// For example now CONSUL_ADDR should be set as CONSUL_HTTP_ADDR.
func NewConsulFromEnv() (*api.Client, error) {
	// for fast approach, if not exist pass
	if _, ok := os.LookupEnv("CONSUL_HTTP_ADDR"); !ok {
		return nil, fmt.Errorf("CONSUL_HTTP_ADDR not exist, err: %w", ErrNoClient)
	}

	return NewConsulWithConfig(api.DefaultConfig())
}

// NewConsulWithConfig creates a client from a config.
func NewConsulWithConfig(config *api.Config) (*api.Client, error) {
	cl, err := api.NewClient(config)

	if cl == nil {
		return nil, ErrNoClient
	}

	return cl, err
}
