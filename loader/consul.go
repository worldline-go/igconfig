package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/consul/api/watch"

	"github.com/rs/zerolog/log"

	"github.com/hashicorp/consul/api"
)

// ConsulConfigPathPrefix specifies prefix for key search.
var ConsulConfigPathPrefix = "finops"

var ErrNoClient = errors.New("no client available")

var _ Loader = Consul{}
var _ DynamicValuer = Consul{}

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
	// By default it is YAML parser.
	//
	// Please prefer YAML to JSON or anything else if there is no strict requirement for it.
	//
	// Note: this function is not used in Watcher.
	Decoder Decoder
}

// Load retrieves data from Consul and decode response into 'to' struct.
func (c Consul) Load(appName string, to interface{}) error {
	if err := c.EnsureClient(); err != nil {
		return err
	}

	data, _, err := c.Client.KV().Get(getConsulConfigPath(appName), nil)
	// If no data or err is returned - return early.
	if data == nil || err != nil {
		return err
	}

	if c.Decoder == nil {
		c.Decoder = DefaultDecoder
	}

	return c.Decoder(bytes.NewReader(data.Value), to)
}

// DynamicValue allows to get dynamically updated values at a runtime.
//
// WARNING: this is experimental feature and is not guaranteed to work. Also it could be changed at will.
//
// ---
//
// If specified key has new value(or was deleted) - runner will be called.
//
// This function will spin up Goroutine to track changes in background, while this function will still be blocking.
// Reason is to be able to track errors returned from it.
//
// Developers will use Goroutine(meaning that total of 2 goroutines will be created)
// when calling this function so that application will not be blocked.
//
// Note: Runner will be called ONLY when new value is received. Removal of the path - still new value,
// and developer MUST handle a case where incoming value is nil.
//
// Example:
//	consul, _ := NewConsulFromEnv()
//
//	// This is the variable that will change it's value dynamically based on Consul value.
//	var externalVar string
//	updateHandler := func(input []byte) error {
//		if input == nil {
//			// This is just an example. This will be called when config field in Consul will be deleted.
//			externalVarDisabled()
//			return nil
//		}
//
//		externalVar = string(input)
//		someOtherHandler() // External handler that maybe should be called to do something.
//	}
//
//	// This context will handle cancellation for DynamicUpdate.
//	// Meaning that when this context will be canceled - DynamicValue will also be stopped.
//	ctx, cancel := context.WithCancel(context.Background)
//	defer cancel()
//
//	go func() {
//		for {
//			err := consul.DynamicValue(ctx, DynamicConfig{
//				AppName: "appName",
//				FieldName: "fieldName", // This can also be sub-key: 'struct/inner/field'
//				Runner: updateHandler,
//			}
//		}
//	}()
//
func (c Consul) DynamicValue(ctx context.Context, config DynamicConfig) error {
	if err := c.EnsureClient(); err != nil {
		return err
	}

	plan, err := watch.Parse(map[string]interface{}{
		"type": "key",
		"key":  getConsulConfigPath(config.AppName, config.FieldName),
	})
	if err != nil {
		return err
	}

	watchCtx, stopWatcher := context.WithCancel(ctx)
	defer func() {
		stopWatcher()
		plan.Stop()
	}()

	var handlerErr error

	plan.HybridHandler = func(_ watch.BlockingParamVal, raw interface{}) {
		var data []byte

		if raw != nil { // nil is a valid return value
			v, ok := raw.(*api.KVPair)
			if ok {
				data = v.Value
			} else {
				// Just to be safe
				handlerErr = fmt.Errorf("unknown dynamic value type received: %T", raw)

				stopWatcher()

				return
			}
		}

		if execErr := executeRunner(config.FieldName, data, config.Runner); execErr != nil {
			handlerErr = execErr

			stopWatcher()
		}
	}

	chanRun := func() <-chan error {
		var ch = make(chan error)

		go func() {
			ch <- plan.RunWithClientAndHclog(c.Client, hclog.NewNullLogger())
		}()

		return ch
	}

	select {
	case <-watchCtx.Done():
		plan.Stop()

		if handlerErr != nil {
			return handlerErr
		}

		return watchCtx.Err()
	case err := <-chanRun():
		return err
	}
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

func NewConsul(addr string) (*api.Client, error) {
	return NewConsulWithConfig(&api.Config{Address: addr})
}

// NewConsulFromEnv creates a client from environmental variables.
//
// This function uses api.DefatulConfig(), which means that variables should be named as Consul expects them.
// For example now CONSUL_ADDR should be set as CONSUL_HTTP_ADDR.
func NewConsulFromEnv() (*api.Client, error) {
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

func executeRunner(keyPath string, newValue []byte, runner DynamicRunner) error {
	if l := log.Debug(); l.Enabled() {
		l.Str("key_path", keyPath).Msg("new dynamic value received")
	}

	return runner(newValue)
}
