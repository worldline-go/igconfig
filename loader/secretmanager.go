package loader

import (
	"bytes"
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/rs/zerolog/log"

	"github.com/worldline-go/igconfig/codec"
)

// SecretManagerProjectIDEnv is the environment variable that enables the Secret Manager loader.
// If this variable is not set, the loader is silently skipped.
const SecretManagerProjectIDEnv = "GCP_PROJECT_ID" //nolint:gosec // env var name, not a credential

// SecretManagerTag is the struct tag used for field name resolution.
var SecretManagerTag = "secret"

// SecretManagerAdditionalSecrets lists secret names loaded before the app-specific secret.
// This mirrors the Vault loader's VaultSecretAdditionalPaths for generic/shared secrets.
var SecretManagerAdditionalSecrets = []string{"generic"}

var _ Loader = &SecretManager{}

// SecretManager loads configuration from GCP Secret Manager.
//
// It is the GCP equivalent of the Vault loader: it fetches a YAML secret
// whose name matches appName and decodes it into the config struct using secret tags.
//
// Loading order (same pattern as Vault):
//  1. Each secret in SecretManagerAdditionalSecrets (default: "generic")
//  2. The app-specific secret named appName
//
// The loader is silently skipped when:
//   - GCP_PROJECT_ID environment variable is not set
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
	// ProjectID is the GCP project ID. If empty, read from GCP_PROJECT_ID env var.
	ProjectID string
}

// LoadWithContext retrieves secrets from GCP Secret Manager and decodes them into 'to'.
// Additional secrets (e.g., "generic") are loaded first; the app-specific secret is loaded last.
func (l *SecretManager) LoadWithContext(ctx context.Context, appName string, to any) error {
	err := l.EnsureClient(ctx)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("SecretManager: client setup failed")

		return err
	}

	for _, name := range SecretManagerAdditionalSecrets {
		err = l.loadSecret(ctx, name, to)
		if err != nil {
			return err
		}
	}

	return l.loadSecret(ctx, appName, to)
}

// Load is the same as LoadWithContext without context.
func (l *SecretManager) Load(appName string, to any) error {
	return l.LoadWithContext(context.Background(), appName, to)
}

// EnsureClient creates and sets a GCP Secret Manager client if needed.
// Returns ErrNoClient if GCP_PROJECT_ID is not set or if the client cannot be created.
func (l *SecretManager) EnsureClient(ctx context.Context) error {
	if l.ProjectID == "" {
		l.ProjectID = os.Getenv(SecretManagerProjectIDEnv)
	}

	if l.ProjectID == "" {
		return fmt.Errorf("%w: %s not set", ErrNoClient, SecretManagerProjectIDEnv)
	}

	if l.Client != nil {
		return nil
	}

	var err error

	l.Client, err = secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: create secret manager client: %w", ErrNoClient, err)
	}

	return nil
}

// loadSecret fetches and decodes a single secret version. Returns nil if the secret does not exist.
func (l *SecretManager) loadSecret(ctx context.Context, secretID string, to any) error {
	resourceName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", l.ProjectID, gcpResourceName(secretID))
	log.Ctx(ctx).Info().Str("resource", resourceName).Msg("SecretManager: fetching secret")

	result, err := l.Client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	})
	if err != nil {
		if isGCPNotFound(err) {
			log.Ctx(ctx).Warn().Str("resource", resourceName).Msg("SecretManager: secret not found, skipping")

			return nil
		}

		return fmt.Errorf("SecretManager.loadSecret %q: %w", secretID, err)
	}

	payload := result.GetPayload().GetData()
	log.Ctx(ctx).Info().Int("bytes", len(payload)).Str("secret", secretID).Msg("SecretManager: received payload")

	err = codec.LoadReaderWithDecoder(bytes.NewReader(payload), to, codec.YAML{}, SecretManagerTag)
	if err != nil {
		return fmt.Errorf("SecretManager.loadSecret %q: %w", secretID, err)
	}

	return nil
}
