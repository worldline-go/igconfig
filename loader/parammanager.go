package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	parametermanager "cloud.google.com/go/parametermanager/apiv1"
	"cloud.google.com/go/parametermanager/apiv1/parametermanagerpb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/worldline-go/igconfig/codec"
)

// ParameterManagerProjectIDEnv is the environment variable that enables the Parameter Manager loader.
// If this variable is not set (and ProjectID is empty), the loader is silently skipped.
// It takes priority over the ProjectID field.
const ParameterManagerProjectIDEnv = "GCP_PROJECT_ID"

// ParameterManagerLocationEnv is the environment variable that sets the GCP location
// (region) Parameter Manager resources are read from. If not set, defaults to "global".
// It takes priority over the Location field.
//
// Some GCP organizations restrict resource creation to specific regions via the
// constraints/gcp.resourceLocations org policy, which does not always allow "global" —
// in that case, parameters must be created in an allowed region and this env var (or the
// Location field) must be set to match.
const ParameterManagerLocationEnv = "GCP_PARAMETER_LOCATION"

// ParameterManagerVersionEnv is the environment variable that sets the parameter version to render.
// If not set, defaults to "latest". It takes priority over the Version field.
const ParameterManagerVersionEnv = "GCP_PARAMETER_VERSION"

// ParameterManagerFallbackEnv is the environment variable (bool) that enables or disables
// falling back to the latest enabled version when "latest" is disabled.
// It takes priority over the DisableVersionFallback field.
const ParameterManagerFallbackEnv = "GCP_PARAMETER_FALLBACK"

// ParameterManagerDefaultLocation is used when neither Location nor the
// ParameterManagerLocationEnv environment variable is set.
const ParameterManagerDefaultLocation = GCPDefaultLocation

// ParameterManagerTag is the struct tag used for field name resolution.
var ParameterManagerTag = "cfg"

var _ Loader = &ParameterManager{}

// ParameterManager loads configuration from GCP Parameter Manager.
//
// It is the GCP equivalent of the Consul loader: it fetches a YAML parameter
// whose name matches appName and decodes it into the config struct using cfg tags.
//
// Environment variables are read on every load and take priority over the struct fields.
//
// The loader is silently skipped when:
//   - GCP_PROJECT_ID environment variable is not set and ProjectID is empty
//   - The GCP client cannot be created (e.g., no Workload Identity or credentials)
//   - The parameter named appName does not exist
//
// Example usage with custom loaders:
//
//	loaders := []loader.Loader{
//	    &loader.Default{},
//	    &loader.Consul{},
//	    &loader.ParameterManager{ProjectID: "my-gcp-project", Location: "europe-west3"},
//	    &loader.Vault{},
//	    &loader.File{},
//	    &loader.Env{},
//	}
//	igconfig.LoadWithLoaders("myapp", &cfg, loaders...)
type ParameterManager struct {
	// Client is the GCP Parameter Manager client. Created automatically from
	// Application Default Credentials if nil.
	Client *parametermanager.Client
	// ProjectID is the GCP project ID.
	//  - GCP_PROJECT_ID env var takes priority over this field.
	ProjectID string
	// Location is the GCP region parameters are read from, default is "global".
	//  - GCP_PARAMETER_LOCATION env var takes priority over this field.
	//
	// Non-"global" locations use the regional Parameter Manager endpoint
	// (parametermanager.<location>.rep.googleapis.com) — the global endpoint only serves
	// "global" resources and returns PERMISSION_DENIED for regional ones.
	Location string
	// Version is the parameter version to render, default is "latest".
	//  - GCP_PARAMETER_VERSION env var takes priority over this field.
	Version string
	// DisableVersionFallback disables falling back to the most recently created enabled
	// version when the "latest" version is disabled. Fallback only applies to "latest".
	//  - GCP_PARAMETER_FALLBACK env var takes priority over this field.
	DisableVersionFallback bool
}

type parameterManagerConfig struct {
	ProjectID string
	Location  string
	Version   string
	Fallback  bool
}

// LoadWithContext retrieves a parameter version from GCP Parameter Manager and decodes it into 'to'.
// If the latest version is disabled, it falls back to the most recently created enabled version.
func (l *ParameterManager) LoadWithContext(ctx context.Context, appName string, to any) error {
	cfg, err := l.config()
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("ParameterManager: config failed")

		return err
	}

	parameterID := gcpResourceName(appName)
	if parameterID == "" {
		return fmt.Errorf("%w: parameter name is empty", ErrNoClient)
	}

	if err := l.ensureClient(ctx, cfg.Location); err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("ParameterManager: client setup failed")

		return err
	}

	paramName := fmt.Sprintf("projects/%s/locations/%s/parameters/%s", cfg.ProjectID, cfg.Location, parameterID)
	resourceName := paramName + "/versions/" + cfg.Version
	log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("ParameterManager: fetching parameter")

	result, err := l.Client.RenderParameterVersion(ctx, &parametermanagerpb.RenderParameterVersionRequest{
		Name: resourceName,
	})
	if err != nil {
		if isGCPNotFound(err) {
			log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("ParameterManager: parameter not found, skipping")

			return nil
		}

		if cfg.Fallback && cfg.Version == GCPDefaultVersion && isGCPFailedPrecondition(err) {
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
	cfg, err := l.config()
	if err != nil {
		return err
	}

	return l.ensureClient(ctx, cfg.Location)
}

// config resolves the loader settings, environment variables take priority over fields.
func (l *ParameterManager) config() (parameterManagerConfig, error) {
	cfg := parameterManagerConfig{
		ProjectID: envOr(ParameterManagerProjectIDEnv, l.ProjectID),
		Location:  envOr(ParameterManagerLocationEnv, l.Location, ParameterManagerDefaultLocation),
		Version:   envOr(ParameterManagerVersionEnv, l.Version, GCPDefaultVersion),
		Fallback:  !l.DisableVersionFallback,
	}

	if cfg.ProjectID == "" {
		return cfg, fmt.Errorf("%w: %s not set", ErrNoClient, ParameterManagerProjectIDEnv)
	}

	if v := envOr(ParameterManagerFallbackEnv); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return cfg, fmt.Errorf("ParameterManager: invalid %s value %q: %w", ParameterManagerFallbackEnv, v, err)
		}

		cfg.Fallback = b
	}

	return cfg, nil
}

func (l *ParameterManager) ensureClient(ctx context.Context, location string) error {
	if l.Client != nil {
		return nil
	}

	var opts []option.ClientOption

	if location != ParameterManagerDefaultLocation {
		// Regional resources are only served from the regional endpoint — the global
		// endpoint returns PERMISSION_DENIED for them.
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("parametermanager.%s.rep.googleapis.com:443", location)))
	}

	client, err := parametermanager.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("%w: create parameter manager client: %w", ErrNoClient, err)
	}

	l.Client = client

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
