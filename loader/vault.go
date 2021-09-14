package loader

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/codec"
	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

var VaultSecretTag = "secret"

// VaultRoleIDEnv specifies the name of environment a variable that holds Vault role id to authenticate with.
const VaultRoleIDEnv = "VAULT_ROLE_ID"

// VaultRoleSecretEnv specifies the name of environment a variable that holds Vault role secret.
const VaultRoleSecretEnv = "VAULT_ROLE_SECRET" // nolint:gosec // false-positive

// VaultSecretBasePath is the base path for secrets.
var VaultSecretBasePath = "finops/data"

// VaultSecretGenericPath is path for storing generic secrets.
// Such generic secrets might be re-used by any number of applications.
// They must not be application specific!
var VaultSecretGenericPath = "generic"

// VaultAppRoleBasePath is the base path for
// app role authentication.
var VaultAppRoleBasePath = "auth/approle/login"

type Vaulter interface {
	Read(path string) (*api.Secret, error)
}

// AuthOption options for authentication.
type AuthOption func(*api.Client) error

// Vault loads secret values from Vault instance.
//
// Generic secrets can also be loaded by using LoadGeneric method
//
// Example usage:
//
//  var config Config // some Config struct
//
//  cl, err := api.NewClient(&api.Config{Address: "http://vault:8200"})
//  if err != nil { ... }
//
//  cl.SetToken("some_token") // this could be also other means of authentication.
//
//  vaultLoader := Vault{Client: cl}
//  err = vaultLoader.Load("adm0001s", &config)
//  if err != nil { ... }
//
//  // config is now populated from Vault.
type Vault struct {
	Client Vaulter
	// ErrOnUnsetable specifies behavior when field is unsetable.
	// If it is true - error will be returned, false - it will be just logged.
	ErrOnUnsetable bool
}

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
func NewVaulterFromEnv() (Vaulter, error) {
	vault, err := api.NewClient(api.DefaultConfig())
	if err != nil || vault == nil {
		return nil, err
	}

	roleID, roleSecret := os.Getenv(VaultRoleIDEnv), os.Getenv(VaultRoleSecretEnv)
	// Check only roleID as roleSecret can be empty in some cases.
	if roleID != "" {
		// Unset previous token to prevent any problems.
		vault.ClearToken()

		if err := SetAppRole(roleID, roleSecret)(vault); err != nil {
			return nil, err
		}
	}

	return NewVaulterFromClient(vault)
}

// NewVaulterFromClient will create a Vaulter client based on input client.
//
// This function will try to choose live Vault instance from the Consul.
func NewVaulterFromClient(cl *api.Client) (Vaulter, error) {
	if cl == nil {
		return nil, ErrNoClient
	}

	if err := FetchVaultAddrFromConsul(cl, (&Consul{}).SearchLiveServices); err != nil {
		return nil, fmt.Errorf("fetch Vault addr from Consul: %w", err)
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
//  var config Config // some Config struct
//
//  vaultLoader, NewVaulterer(addr, SetAppRole(roleID, ""))
//  if err != nil { ... }
//
//  loader.VaultSecretBasePath = "some/path/"
//
//  err = vaultLoader.Load("adm0001s", &config)
//  if err != nil { ... }
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
		resp, err := c.Logical().Write(VaultAppRoleBasePath, map[string]interface{}{
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

// Load will load data from Vault to input struct 'to'.
// 'name' is base secret path, or just name of application.
// Path will be constructed as "${VaultSecretTag}/${name}".
// By default VaultSecretTag value is "secrets/data", which allows to load secrets from root.
func (v *Vault) Load(appName string, to interface{}) error {
	if err := v.LoadGeneric(to); err != nil {
		return err
	}

	return v.LoadFromReformat(getVaultSecretPath(appName), to)
}

// LoadGeneric loads generic(shared) secrets from Vault.
func (v *Vault) LoadGeneric(to interface{}) error {
	return v.LoadFromReformat(getVaultSecretPath(VaultSecretGenericPath), to)
}

func (v *Vault) LoadFromReformat(appName string, to interface{}) error {
	secretMap, err := v.loadSecretData(appName)
	if err != nil {
		return err
	}

	return codec.MapDecoder(secretMap, to, VaultSecretTag)
}

func (v *Vault) loadSecretData(appName string) (map[string]interface{}, error) {
	if err := v.EnsureClient(); err != nil {
		return nil, err
	}

	pathSecret, err := v.Client.Read(appName)
	if err != nil {
		return nil, fmt.Errorf("vault request for path %q failed: %w", appName, err)
	}

	if pathSecret == nil || pathSecret.Data == nil {
		// Create dummy data so we will be able to proceed on normal route.
		// We do this as inner structs still may be able to have secret data.
		pathSecret = &api.Secret{
			Data: map[string]interface{}{
				"data": map[string]interface{}{},
			},
		}
	}

	// Empty assign is so this part will not panic if conversion was unsuccessful.
	secretMap, _ := pathSecret.Data["data"].(map[string]interface{})
	if secretMap == nil {
		return nil, fmt.Errorf("secret from path %q cannot be converted to map", appName)
	}

	return secretMap, nil
}

// EnsureClient creates and sets a Vault client if needed.
func (v *Vault) EnsureClient() error {
	if v.Client == nil {
		var err error

		v.Client, err = NewVaulterFromEnv()
		if err != nil {
			return err
		}
	}

	if v.Client == nil {
		return ErrNoClient
	}

	return nil
}

// FetchVaultAddrFromConsul will try to find alive Vault instances in `serviceFetcher`
// and will set `client` address to random instance.
// If no instances were found or error happened - nothing will be changed.
//
// If address will be changed - it will always be HTTPS.
func FetchVaultAddrFromConsul(client *api.Client, serviceFetcher LiveServiceFetcher) error {
	services, err := serviceFetcher(context.Background(), "vault", nil)
	if err != nil {
		if internal.IsLocalNetworkError(err) {
			log.Warn().
				Str("loader", fmt.Sprintf("%T", (*Vault)(nil))).
				Msg("local Consul server is not available, skipping fetching Vault address")

			return nil
		}

		return fmt.Errorf("fetch services: %w", err)
	}

	if len(services) == 0 {
		log.Warn().Msg("no healthy Vault services found in Consul, will keep current address")

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

	return nil
}

func getVaultSecretPath(parts ...string) string {
	return path.Join(append([]string{VaultSecretBasePath}, parts...)...)
}
