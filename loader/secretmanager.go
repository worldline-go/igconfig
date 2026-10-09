package loader

import (
	"bytes"
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"

	"github.com/worldline-go/igconfig/codec"
)

// SecretManagerProjectIDEnv is the environment variable that enables the Secret Manager loader.
// If this variable is not set (and ProjectID is empty), the loader is silently skipped.
// It takes priority over the ProjectID field.
const SecretManagerProjectIDEnv = "GCP_PROJECT_ID" //nolint:gosec // env var name, not a credential

// SecretManagerLocationEnv is the environment variable that sets the GCP location
// (region) secrets are read from. If not set, defaults to "global".
// It takes priority over the Location field.
const SecretManagerLocationEnv = "GCP_SECRET_LOCATION" //nolint:gosec // env var name, not a credential

// SecretManagerVersionEnv is the environment variable that sets the secret version to access.
// If not set, defaults to "latest". It takes priority over the Version field.
const SecretManagerVersionEnv = "GCP_SECRET_VERSION" //nolint:gosec // env var name, not a credential

// SecretManagerAdditionalEnv is the environment variable (comma-separated) that sets the
// additional secrets loaded before the app-specific secret.
// It takes priority over the AdditionalSecrets field and SecretManagerAdditionalSecrets.
const SecretManagerAdditionalEnv = "GCP_SECRET_ADDITIONAL" //nolint:gosec // env var name, not a credential

// SecretManagerTag is the struct tag used for field name resolution.
var SecretManagerTag = "secret"

// SecretManagerAdditionalSecrets lists secret names loaded before the app-specific secret.
// This mirrors the Vault loader's VaultSecretAdditionalPaths for generic/shared secrets.
// Used when neither GCP_SECRET_ADDITIONAL nor the AdditionalSecrets field is set.
var SecretManagerAdditionalSecrets = []string{"generic"}

var _ Loader = &SecretManager{}

// SecretManager loads configuration from GCP Secret Manager.
//
// It is the GCP equivalent of the Vault loader: it fetches a YAML secret
// whose name matches appName and decodes it into the config struct using secret tags.
//
// Loading order (same pattern as Vault):
//  1. Each additional secret (default: "generic")
//  2. The app-specific secret named appName
//
// Environment variables are read on every load and take priority over the struct fields.
//
// The loader is silently skipped when:
//   - GCP_PROJECT_ID environment variable is not set and ProjectID is empty
//   - The GCP client cannot be created (e.g., no Workload Identity or credentials)
//
// Individual secrets that do not exist are silently skipped (nil return),
// matching the Vault loader's behavior for missing paths.
//
// Example usage with custom loaders:
//
//	loaders := []loader.Loader{
//	    &loader.Default{},
//	    &loader.Consul{},
//	    &loader.ParameterManager{ProjectID: "my-gcp-project"},
//	    &loader.Vault{},
//	    &loader.SecretManager{ProjectID: "my-gcp-project"},
//	    &loader.File{},
//	    &loader.Env{},
//	}
//	igconfig.LoadWithLoaders("myapp", &cfg, loaders...)
type SecretManager struct {
	// Client is the GCP Secret Manager client. Created automatically from
	// Application Default Credentials if nil.
	Client *secretmanager.Client
	// ProjectID is the GCP project ID.
	//  - GCP_PROJECT_ID env var takes priority over this field.
	ProjectID string
	// Location is the GCP region secrets are read from, default is "global".
	//  - GCP_SECRET_LOCATION env var takes priority over this field.
	//
	// Non-"global" locations use the regional Secret Manager endpoint
	// (secretmanager.<location>.rep.googleapis.com).
	Location string
	// Version is the secret version to access, default is "latest".
	//  - GCP_SECRET_VERSION env var takes priority over this field.
	Version string
	// AdditionalSecrets are loaded before the app-specific secret, missing ones are skipped.
	// If nil, SecretManagerAdditionalSecrets is used.
	//  - GCP_SECRET_ADDITIONAL env var (comma-separated) takes priority over this field.
	AdditionalSecrets []string
}

type secretManagerConfig struct {
	ProjectID  string
	Location   string
	Version    string
	Additional []string
}

// LoadWithContext retrieves secrets from GCP Secret Manager and decodes them into 'to'.
// Additional secrets (e.g., "generic") are loaded first; the app-specific secret is loaded last.
func (l *SecretManager) LoadWithContext(ctx context.Context, appName string, to any) error {
	cfg, err := l.config()
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("SecretManager: config failed")

		return err
	}

	if err := l.ensureClient(ctx, cfg.Location); err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("SecretManager: client setup failed")

		return err
	}

	for _, name := range cfg.Additional {
		if err := l.loadSecret(ctx, cfg, name, to); err != nil {
			return err
		}
	}

	return l.loadSecret(ctx, cfg, appName, to)
}

// Load is the same as LoadWithContext without context.
func (l *SecretManager) Load(appName string, to any) error {
	return l.LoadWithContext(context.Background(), appName, to)
}

// EnsureClient creates and sets a GCP Secret Manager client if needed.
// Returns ErrNoClient if GCP_PROJECT_ID is not set or if the client cannot be created.
func (l *SecretManager) EnsureClient(ctx context.Context) error {
	cfg, err := l.config()
	if err != nil {
		return err
	}

	return l.ensureClient(ctx, cfg.Location)
}

// config resolves the loader settings, environment variables take priority over fields.
func (l *SecretManager) config() (secretManagerConfig, error) {
	cfg := secretManagerConfig{
		ProjectID: envOr(SecretManagerProjectIDEnv, l.ProjectID),
		Location:  envOr(SecretManagerLocationEnv, l.Location, GCPDefaultLocation),
		Version:   envOr(SecretManagerVersionEnv, l.Version, GCPDefaultVersion),
	}

	if cfg.ProjectID == "" {
		return cfg, fmt.Errorf("%w: %s not set", ErrNoClient, SecretManagerProjectIDEnv)
	}

	switch {
	case envOr(SecretManagerAdditionalEnv) != "":
		cfg.Additional = splitComma(envOr(SecretManagerAdditionalEnv))
	case l.AdditionalSecrets != nil:
		cfg.Additional = l.AdditionalSecrets
	default:
		cfg.Additional = SecretManagerAdditionalSecrets
	}

	return cfg, nil
}

func (l *SecretManager) ensureClient(ctx context.Context, location string) error {
	if l.Client != nil {
		return nil
	}

	var opts []option.ClientOption

	if location != GCPDefaultLocation {
		// Regional secrets are only served from the regional endpoint.
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("secretmanager.%s.rep.googleapis.com:443", location)))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("%w: create secret manager client: %w", ErrNoClient, err)
	}

	l.Client = client

	return nil
}

// secretResourceName constructs the full secret version resource name.
func secretResourceName(cfg secretManagerConfig, secretID string) string {
	if cfg.Location == GCPDefaultLocation {
		return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", cfg.ProjectID, secretID, cfg.Version)
	}

	return fmt.Sprintf("projects/%s/locations/%s/secrets/%s/versions/%s", cfg.ProjectID, cfg.Location, secretID, cfg.Version)
}

// loadSecret fetches and decodes a single secret version. Returns nil if the secret does not exist.
func (l *SecretManager) loadSecret(ctx context.Context, cfg secretManagerConfig, name string, to any) error {
	secretID := gcpResourceName(name)
	if secretID == "" {
		return nil
	}

	resourceName := secretResourceName(cfg, secretID)
	log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("SecretManager: fetching secret")

	result, err := l.Client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	})
	if err != nil {
		if isGCPNotFound(err) {
			log.Ctx(ctx).Debug().Str("resource", resourceName).Msg("SecretManager: secret not found, skipping")

			return nil
		}

		return fmt.Errorf("SecretManager.loadSecret %q: %w", name, err)
	}

	payload := result.GetPayload().GetData()
	log.Ctx(ctx).Debug().Int("bytes", len(payload)).Msg("SecretManager: received payload")

	err = codec.LoadReaderWithDecoder(bytes.NewReader(payload), to, codec.YAML{}, SecretManagerTag)
	if err != nil {
		return fmt.Errorf("SecretManager.loadSecret %q: %w", name, err)
	}

	return nil
}
