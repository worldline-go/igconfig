package loader

import (
	"bytes"
	"context"
	"fmt"
	"os"

	parametermanager "cloud.google.com/go/parametermanager/apiv1"
	"cloud.google.com/go/parametermanager/apiv1/parametermanagerpb"

	"github.com/worldline-go/igconfig/codec"
)

// ParameterManagerProjectIDEnv is the environment variable that enables the Parameter Manager loader.
// If this variable is not set, the loader is silently skipped.
const ParameterManagerProjectIDEnv = "GCP_PROJECT_ID"

// ParameterManagerTag is the struct tag used for field name resolution.
var ParameterManagerTag = "cfg"

var _ Loader = &ParameterManager{}

// ParameterManager loads configuration from GCP Parameter Manager.
//
// It is the GCP equivalent of the Consul loader: it fetches a YAML parameter
// whose name matches appName and decodes it into the config struct using cfg tags.
//
// The loader is silently skipped when:
//   - GCP_PROJECT_ID environment variable is not set
//   - The GCP client cannot be created (e.g., no Workload Identity or credentials)
//   - The parameter named appName does not exist
//
// Example usage with custom loaders:
//
//	loaders := []loader.Loader{
//	    &loader.Default{},
//	    &loader.Consul{},
//	    &loader.ParameterManager{ProjectID: "my-gcp-project"},
//	    &loader.Vault{},
//	    &loader.File{},
//	    &loader.Env{},
//	}
//	igconfig.LoadWithLoaders("myapp", &cfg, loaders...)
type ParameterManager struct {
	// Client is the GCP Parameter Manager client. Created automatically from
	// Application Default Credentials if nil.
	Client *parametermanager.Client
	// ProjectID is the GCP project ID. If empty, read from GCP_PROJECT_ID env var.
	ProjectID string
}

// LoadWithContext retrieves a parameter version from GCP Parameter Manager and decodes it into 'to'.
func (l *ParameterManager) LoadWithContext(ctx context.Context, appName string, to interface{}) error {
	if err := l.EnsureClient(ctx); err != nil {
		return err
	}

	result, err := l.Client.RenderParameterVersion(ctx, &parametermanagerpb.RenderParameterVersionRequest{
		Name: fmt.Sprintf("projects/%s/locations/global/parameters/%s/versions/latest", l.ProjectID, appName),
	})
	if err != nil {
		if isGCPNotFound(err) {
			return nil
		}

		return fmt.Errorf("ParameterManager.LoadWithContext: %w", err)
	}

	if err := codec.LoadReaderWithDecoder(bytes.NewReader(result.RenderedPayload), to, codec.YAML{}, ParameterManagerTag); err != nil {
		return fmt.Errorf("ParameterManager.LoadWithContext: %w", err)
	}

	return nil
}

// Load is the same as LoadWithContext without context.
func (l *ParameterManager) Load(appName string, to interface{}) error {
	return l.LoadWithContext(context.Background(), appName, to)
}

// EnsureClient creates and sets a GCP Parameter Manager client if needed.
// Returns ErrNoClient if GCP_PROJECT_ID is not set or if the client cannot be created.
func (l *ParameterManager) EnsureClient(ctx context.Context) error {
	if l.ProjectID == "" {
		l.ProjectID = os.Getenv(ParameterManagerProjectIDEnv)
	}

	if l.ProjectID == "" {
		return fmt.Errorf("%w: %s not set", ErrNoClient, ParameterManagerProjectIDEnv)
	}

	if l.Client != nil {
		return nil
	}

	var err error

	l.Client, err = parametermanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: create parameter manager client: %v", ErrNoClient, err)
	}

	return nil
}
