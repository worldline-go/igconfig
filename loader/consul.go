package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/codec"

	"github.com/hashicorp/consul/api"
)

var ConsulTag = "cfg"

// ConsulConfigPathPrefix specifies prefix for key search.
var ConsulConfigPathPrefix = "finops"

var ErrNoClient = errors.New("no client available")

var _ Loader = Consul{}

var _ DynamicValuer = Consul{}

// LiveServiceFetcher is a signature of the function that will fetch only live instances of the service.
//
// If no services found - (nil, nil) will be returned.
type LiveServiceFetcher func(ctx context.Context, name string, tags []string) ([]*api.ServiceEntry, error)

type Consuler interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
}

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
func (c Consul) LoadWithContext(ctx context.Context, appName string, to interface{}) error {
	if err := c.EnsureClient(); err != nil {
		return err
	}

	queryOptions := api.QueryOptions{}
	data, _, err := c.Client.KV().Get(getConsulConfigPath(appName), queryOptions.WithContext(ctx))
	// If no data or err is returned - return early.
	if data == nil || err != nil {
		return err
	}

	if c.Decoder == nil {
		c.Decoder = codec.YAML{}
	}

	if err := codec.LoadReaderWithDecoder(bytes.NewReader(data.Value), to, c.Decoder, ConsulTag); err != nil {
		return fmt.Errorf("Consul.LoadWithContext error: %w", err)
	}

	return nil
}

// Load is just same as LoadWithContext without context.
func (c Consul) Load(appName string, to interface{}) error {
	return c.LoadWithContext(context.Background(), appName, to)
}

// EnsureClient creates and sets a Consul client if needed.
func (c *Consul) EnsureClient() error {
	if c.Client == nil {
		var err error

		c.Client, err = NewConsulFromEnv()
		if err != nil {
			return err
		}
	}

	if c.Client == nil {
		return ErrNoClient
	}

	return nil
}

// SearchLiveServices is a wrapper for c.Client.Health().ServiceMultipleTags(name, tags, true, (&api.QueryOptions{}).WithContext(ctx))
//
// This provides a bit nicer interface on fetching services
// and gives ability to have LiveServiceFetcher as an argument or a field instead of actual implementation.
func (c *Consul) SearchLiveServices(ctx context.Context, name string, tags []string) ([]*api.ServiceEntry, error) {
	if err := c.EnsureClient(); err != nil {
		return nil, err
	}

	services, _, err := c.Client.Health().ServiceMultipleTags(name, tags, true, (&api.QueryOptions{}).WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("fetch service instances: %w", err)
	}

	return services, nil
}

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

func NewConsulWithConfig(config *api.Config) (*api.Client, error) {
	cl, err := api.NewClient(config)

	if cl == nil {
		return nil, ErrNoClient
	}

	return cl, err
}

func getConsulConfigPath(parts ...string) string {
	return path.Join(append([]string{ConsulConfigPathPrefix}, parts...)...)
}
