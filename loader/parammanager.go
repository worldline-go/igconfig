package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	parametermanager "cloud.google.com/go/parametermanager/apiv1"
	"cloud.google.com/go/parametermanager/apiv1/parametermanagerpb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"

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
// If the latest version is disabled, it falls back to the most recently created enabled version.
func (l *ParameterManager) LoadWithContext(ctx context.Context, appName string, to any) error {
	err := l.EnsureClient(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("ParameterManager: client setup failed")

		return err
	}

	paramName := fmt.Sprintf("projects/%s/locations/global/parameters/%s", l.ProjectID, gcpResourceName(appName))
	resourceName := paramName + "/versions/latest"
	log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("ParameterManager: fetching parameter")

	result, err := l.Client.RenderParameterVersion(ctx, &parametermanagerpb.RenderParameterVersionRequest{
		Name: resourceName,
	})
	if err != nil {
		if isGCPNotFound(err) {
			log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("ParameterManager: parameter not found, skipping")

			return nil
		}

		if isGCPFailedPrecondition(err) {
			log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("ParameterManager: latest version disabled, searching for latest enabled version")

			return l.loadLatestEnabledVersion(ctx, paramName, to)
		}

		return fmt.Errorf("ParameterManager.LoadWithContext: %w", err)
	}

	payload := result.GetRenderedPayload()
	log.Ctx(ctx).Debug().Int("bytes", len(payload)).Msg("ParameterManager: received payload")

	err = codec.LoadReaderWithDecoder(bytes.NewReader(payload), to, codec.YAML{}, ParameterManagerTag)
	if err != nil {
		return fmt.Errorf("ParameterManager.LoadWithContext: %w", err)
	}

	return nil
}

// Load is the same as LoadWithContext without context.
func (l *ParameterManager) Load(appName string, to any) error {
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
		return fmt.Errorf("%w: create parameter manager client: %w", ErrNoClient, err)
	}

	return nil
}

// loadLatestEnabledVersion lists all versions of a parameter and renders the most recently
// created enabled one. Returns nil (skip) if no enabled versions exist.
func (l *ParameterManager) loadLatestEnabledVersion(ctx context.Context, paramName string, to any) error {
	var latestName string

	var latestTime time.Time

	iter := l.Client.ListParameterVersions(ctx, &parametermanagerpb.ListParameterVersionsRequest{
		Parent: paramName,
	})

	for {
		v, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return fmt.Errorf("ParameterManager.loadLatestEnabledVersion: list versions: %w", err)
		}

		if v.GetDisabled() {
			continue
		}

		if ct := v.GetCreateTime().AsTime(); latestName == "" || ct.After(latestTime) {
			latestName = v.GetName()
			latestTime = ct
		}
	}

	if latestName == "" {
		log.Ctx(ctx).Debug().Str("parameter", paramName).Msg("ParameterManager: no enabled versions found, skipping")

		return nil
	}

	log.Ctx(ctx).Debug().Str("resource", latestName).Msg("ParameterManager: rendering latest enabled version")

	result, err := l.Client.RenderParameterVersion(ctx, &parametermanagerpb.RenderParameterVersionRequest{
		Name: latestName,
	})
	if err != nil {
		return fmt.Errorf("ParameterManager.loadLatestEnabledVersion: %w", err)
	}

	payload := result.GetRenderedPayload()
	log.Ctx(ctx).Debug().Int("bytes", len(payload)).Msg("ParameterManager: received payload")

	err = codec.LoadReaderWithDecoder(bytes.NewReader(payload), to, codec.YAML{}, ParameterManagerTag)
	if err != nil {
		return fmt.Errorf("ParameterManager.loadLatestEnabledVersion: %w", err)
	}

	return nil
}
