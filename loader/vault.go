package loader

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"

	"github.com/worldline-go/igconfig/codec"
	"github.com/worldline-go/igconfig/internal"
)

// AdditionalPath is used to add additional path to the Vault path.
type AdditionalPath struct {
	// Map to wrap values to a new map with this key.
	Map string
	// Name to get that value.
	Name string
}

// VaultSecretTag is the tag used to identify Vault secrets.
var VaultSecretTag = "secret"

// VaultRoleIDEnv specifies the name of environment a variable that holds Vault role id to authenticate with.
const VaultRoleIDEnv = "VAULT_ROLE_ID"

// VaultRoleSecretEnv specifies the name of environment a variable that holds Vault role secret.
const VaultRoleSecretEnv = "VAULT_ROLE_SECRET" // nolint:gosec // false-positive

// VaultAppRoleBasePathEnv specifies the name of the environment variable that holds the path to login.
const VaultAppRoleBasePathEnv = "VAULT_APP_ROLE_BASE_PATH"

// VaultSecretBasePathEnv specifies the name of the environment variable that holds the base path to secrets.
const VaultSecretBasePathEnv = "VAULT_SECRET_BASE_PATH"

// VaultSecretDefaultBasePath is the base path for secrets.
const VaultSecretDefaultBasePath = "finops"

// VaultAppRoleDefaultBasePath is the default base path for
// app role authentication.
const VaultAppRoleDefaultBasePath = "auth/approle/login"

// VaultSecretAdditionalPaths is usable for additional secrets to load.
// Such generic secrets might be re-used by any number of applications.
// They must not be application specific!
var VaultSecretAdditionalPaths = []AdditionalPath{
	{Map: "", Name: "generic"},
}

var errUnusable = errors.New("method not usable")

// Vaulter interface for Vault.
type Vaulter interface {
	Read(path string) (*api.Secret, error)
	List(path string) (*api.Secret, error)
}

// AuthOption options for authentication.
type AuthOption func(*api.Client) error

// Vault loads secret values from Vault instance.
//
// # Generic secrets can also be loaded by using LoadGeneric method
//
// Example usage:
//
//	var config Config // some Config struct
//
//	cl, err := api.NewClient(&api.Config{Address: "http://vault:8200"})
//	if err != nil { ... }
//
//	cl.SetToken("some_token") // this could be also other means of authentication.
//
//	vaultLoader := Vault{Client: cl}
//	err = vaultLoader.Load("adm0001s", &config)
//	if err != nil { ... }
//
//	// config is now populated from Vault.
type Vault struct {
	Client Vaulter
}

// NewVaulter creates a new Vault client.
func NewVaulter(addr, token string) (Vaulter, error) {
	cl, err := api.NewClient(&api.Config{Address: addr})
	if err == nil {
		cl.SetToken(token)
	}

	if cl == nil {
		return nil, ErrNoClient
	}

	return cl.Logical(), err
}

// NewVaulterFromEnv creates default Vault client.
//
// If VaultRoleIDEnv environment variable is set - also logs in based on role_id and role_secret.
//
// See NewVaulterFromClient for more information.
func NewVaulterFromEnv(ctx context.Context) (Vaulter, error) {
	vault, err := api.NewClient(api.DefaultConfig())
	if err != nil || vault == nil {
		return nil, err
	}

	return NewVaulterFromClient(ctx, vault)
}

// NewVaulterFromClient will create a Vaulter client based on input client.
//
// This function will try to choose live Vault instance from the Consul.
func NewVaulterFromClient(ctx context.Context, cl *api.Client) (Vaulter, error) {
	if cl == nil {
		return nil, ErrNoClient
	}

	err := ErrNoClient
	if v, _ := strconv.ParseBool(os.Getenv("VAULT_CONSUL_ADDR_DISABLE")); !v {
		// Override vault address from consul
		err = FetchVaultAddrFromConsul(ctx, cl, (&Consul{}).SearchLiveServices)
		if err != nil && !errors.Is(err, ErrNoClient) {
			return nil, fmt.Errorf("fetch Vault addr from Consul: %w", err)
		}
	}

	// not get any address with environment value
	if errors.Is(err, ErrNoClient) {
		// check VAULT_ADDR and VAULT_AGENT_ADDR to not change vault address
		if _, ok := os.LookupEnv("VAULT_ADDR"); ok {
			return setAppRoleEnv(cl)
		}

		if _, ok := os.LookupEnv("VAULT_AGENT_ADDR"); ok {
			return setAppRoleEnv(cl)
		}

		return nil, fmt.Errorf("VAULT_ADDR, VAULT_AGENT_ADDR or CONSUL_HTTP_ADDR not found, err: %w", ErrNoClient)
	}

	return setAppRoleEnv(cl)
}

func setAppRoleEnv(cl *api.Client) (Vaulter, error) {
	roleID, roleSecret := os.Getenv(VaultRoleIDEnv), os.Getenv(VaultRoleSecretEnv)
	// Check only roleID as roleSecret can be empty in some cases.
	if roleID != "" {
		// Unset previous token to prevent any problems.
		cl.ClearToken()

		if err := SetAppRole(roleID, roleSecret)(cl); err != nil {
			return nil, err
		}
	}

	return cl.Logical(), nil
}

// SimpleVaultLoad will load secret data based without need to configure everything.
//
// Deprecated: Please try to use NewVaulterFromEnv or NewVaulterFromClient to create a client and call Load on it.
func SimpleVaultLoad(addr, token, name string, to interface{}) error {
	cl, err := NewVaulter(addr, token)
	if err != nil {
		return err
	}

	return (&Vault{Client: cl}).Load(name, to)
}

// NewVaulterer returns the Vaulter interface. If role_id is given it will call SetTokenAppRole
// to set the token
//
// Example usage:
//
//	var config Config // some Config struct
//
//	vaultLoader, NewVaulterer(addr, SetAppRole(roleID, ""))
//	if err != nil { ... }
//
//	err = vaultLoader.Load("adm0001s", &config)
//	if err != nil { ... }
func NewVaulterer(addr string, opts ...AuthOption) (loader Loader, err error) {
	cl, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		if err = opt(cl); err != nil {
			return nil, err
		}
	}

	return &Vault{Client: cl.Logical()}, err
}

// SetToken sets the token auth method
// It does not do authentication here.
func SetToken(token string) AuthOption {
	return func(c *api.Client) error {
		c.SetToken(token)

		return nil
	}
}

// SetAppRole sets the client token from the Approle auth Method
// It does authenticate to fetch the token, then sets it.
//
// Vault can also be setup to authenticate with role_id only
// for this the secret id can be passed as blank.
func SetAppRole(role, secret string) AuthOption {
	return func(c *api.Client) error {
		path := internal.GetEnvWithFallback(VaultAppRoleBasePathEnv, VaultAppRoleDefaultBasePath)
		resp, err := c.Logical().Write(path, map[string]interface{}{
			"role_id":   role,
			"secret_id": secret,
		})
		if err != nil {
			return err
		}

		c.SetToken(resp.Auth.ClientToken)

		return nil
	}
}

// LoadWithContext will load data from Vault to input struct 'to'.
// 'name' is base secret path, or just name of application.
// Path will be constructed as "${VaultSecretTag}/${name}".
// By default VaultSecretTag value is "secrets/data", which allows to load secrets from root.
func (l *Vault) LoadWithContext(ctx context.Context, appName string, to interface{}) error {
	if err := l.LoadGeneric(ctx, to); err != nil {
		return err
	}

	return l.LoadFromReformat(ctx, []AdditionalPath{{Map: "", Name: appName}}, to)
}

// Load is same as LoadWithContext without context.
func (l *Vault) Load(appName string, to interface{}) error {
	return l.LoadWithContext(context.Background(), appName, to)
}

// LoadGeneric loads generic(shared) secrets from Vault.
func (l *Vault) LoadGeneric(ctx context.Context, to interface{}) error {
	return l.LoadFromReformat(ctx, VaultSecretAdditionalPaths, to)
}

// LoadFromReformat loads secrets from Vault and load to the input struct 'to'.
func (l *Vault) LoadFromReformat(ctx context.Context, paths []AdditionalPath, to interface{}) error {
	for _, path := range paths {
		secretMap, err := l.loadSecretData(ctx, path.Name, true)
		if err != nil {
			return err
		}

		if path.Map != "" {
			secretMap = map[string]interface{}{
				path.Map: secretMap,
			}
		}

		if err := codec.MapDecoder(secretMap, to, VaultSecretTag); err != nil {
			//nolint:wrapcheck // not need
			return err
		}
	}

	return nil
}

func (l *Vault) loadSecretData(ctx context.Context, appName string, errCheck bool) (map[string]interface{}, error) {
	if err := l.EnsureClient(ctx); err != nil {
		return nil, err
	}

	// first try to check list method
	if strings.HasSuffix(appName, "/") {
		// list with meta path
		rest, err := l.list(ctx, appName)
		if err == nil {
			return rest, err
		}

		// error from not list area
		if !errors.Is(err, errUnusable) {
			return nil, err
		}
	}

	// read with data path
	rest, err := l.read(ctx, appName, errCheck)
	if err == nil {
		return rest, err
	}

	// second try list
	if errors.Is(err, errUnusable) {
		rest, err := l.list(ctx, appName)
		// dont return errUnusable
		if !errors.Is(err, errUnusable) {
			return rest, err
		}

		return nil, nil
	}

	return rest, err
}

func (l *Vault) list(ctx context.Context, appName string) (map[string]interface{}, error) {
	secretBasePath := internal.GetEnvWithFallback(VaultSecretBasePathEnv, VaultSecretDefaultBasePath)
	appNameMeta := path.Join(secretBasePath, "metadata", appName)

	pathSecret, _ := l.Client.List(appNameMeta)

	if pathSecret != nil {
		// combine new map and return
		secretMap := make(map[string]interface{})

		keys, ok := pathSecret.Data["keys"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("data[\"keys\"] from path %q cannot be converted to array", appName)
		}

		for _, k := range keys {
			data, err := l.loadSecretData(ctx, path.Join(appName, k.(string)), false)
			if err != nil {
				return nil, err
			}

			secretMap[k.(string)] = data
		}

		return secretMap, nil
	}

	return nil, errUnusable
}

func (l *Vault) read(ctx context.Context, appName string, errCheck bool) (map[string]interface{}, error) {
	secretBasePath := internal.GetEnvWithFallback(VaultSecretBasePathEnv, VaultSecretDefaultBasePath)
	appNameData := path.Join(secretBasePath, "data", appName)

	pathSecret, err := l.Client.Read(appNameData)
	if err != nil {
		// recursive call should not return error
		// could be policy denied to read it
		if !errCheck {
			log.Ctx(ctx).Warn().Err(err).Msgf("denied to read path %v failed", appName)

			return nil, nil
		}

		return nil, fmt.Errorf("vault request for path %q failed: %w", appName, err)
	}

	if pathSecret != nil {
		// Is it destroyed?
		metadata, ok := pathSecret.Data["metadata"].(map[string]interface{})
		if ok && isDestroyed(metadata) {
			log.Ctx(ctx).Warn().Msgf("%s is destoyed, skipping", path.Join(secretBasePath, appName))

			return nil, nil
		}

		secretMap, _ := pathSecret.Data["data"].(map[string]interface{})
		if secretMap == nil {
			return nil, fmt.Errorf("secret from path %q cannot be converted to map", appName)
		}

		return secretMap, nil
	}

	return nil, errUnusable
}

// EnsureClient creates and sets a Vault client if needed.
func (l *Vault) EnsureClient(ctx context.Context) error {
	if l.Client == nil {
		var err error

		l.Client, err = NewVaulterFromEnv(ctx)
		if err != nil {
			return err
		}
	}

	if l.Client == nil {
		return ErrNoClient
	}

	return nil
}

// FetchVaultAddrFromConsul will try to find alive Vault instances in `serviceFetcher`
// and will set `client` address to random instance.
// If no instances were found or error happened - nothing will be changed.
//
// If address will be changed - it will always be HTTPS.
func FetchVaultAddrFromConsul(ctx context.Context, client *api.Client, serviceFetcher LiveServiceFetcher) error {
	services, err := serviceFetcher(ctx, "vault", nil)
	if err != nil {
		if internal.IsLocalNetworkError(err) {
			log.Ctx(ctx).Warn().
				Str("loader", fmt.Sprintf("%T", (*Vault)(nil))).
				Msg("local Consul server is not available, skipping fetching Vault address")

			return nil
		}

		return fmt.Errorf("fetch services: %w", err)
	}

	if len(services) == 0 {
		log.Ctx(ctx).Warn().Msg("no healthy Vault services found in Consul, will keep current address")

		return nil
	}
	// newRand is necessary to not overwrite global random source,
	// which could configured in special way.
	// Plus it removes side effects on rand package.
	newRand := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // No security here, just randomizer.
	randService := services[newRand.Intn(len(services))].Service

	if err := client.SetAddress(fmt.Sprintf("https://%s:%d", randService.Address, randService.Port)); err != nil {
		return fmt.Errorf("set address: %w", err)
	}

	log.Ctx(ctx).Info().Msg("vault address got from consul server")

	return nil
}

// isDestroyed checks data is destroyed.
func isDestroyed(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}

	if destoyed, ok := metadata["destroyed"].(bool); ok && destoyed {
		return true
	}

	return false
}
