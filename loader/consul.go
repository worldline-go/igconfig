package loader

import (
	"bytes"
	"io"

	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

// ConsulConfigPathPrefix specifies prefix for key search.
// MUST always end in slash!
var ConsulConfigPathPrefix = "config/"

type Consuler interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
}

// Consul is an instance of configuration loader from Consul.
//
// Example usage:
//
//  var config Config // some Config struct
//
//  cl, err := api.NewClient(&api.Config{Address: "http://consul:8500"})
//  if err != nil { ... }
//
//  consulLoader := Consul{Client: cl}
//  err = consulLoader.Load("adm0001s", &config)
//  if err != nil { ... }
//
//  // config is now populated from Consul.
type Consul struct {
	Client Consuler
	// Decoder specifies function that will decode the response from Consul.
	// By default it is YAML parser.
	//
	// Please prefer YAML to JSON or anything else if there is no strict requirement for it.
	Decoder func(r io.Reader, to interface{}) error
}

func DefaultDecoder(r io.Reader, to interface{}) error {
	return yaml.NewDecoder(r).Decode(to)
}

// Load retrieves data from Consul and decode response into 'to' struct.
func (c Consul) Load(path string, to interface{}) error {
	data, _, err := c.Client.Get(ConsulConfigPathPrefix+path, nil)
	// If no data or err is returned - return early.
	if data == nil || err != nil {
		return err
	}

	if c.Decoder == nil {
		c.Decoder = DefaultDecoder
	}

	return c.Decoder(bytes.NewReader(data.Value), to)
}

func NewConsuler(addr string) (Consuler, error) {
	return NewConsulerWithConfig(&api.Config{Address: addr})
}

func NewConsulerWithConfig(config *api.Config) (Consuler, error) {
	cl, err := api.NewClient(config)

	// Yes, cl may be nil, but this method will not panic even if it is.
	// As such - it is possible to use it in this way.
	return cl.KV(), err
}
