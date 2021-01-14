package loader

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"

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

// SkipStructTypes specifies types of structs that should be skipped from setting.
var SkipStructTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(time.Time{}): {},
}

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

func NewVaulterFromClient(cl *api.Client) (Vaulter, error) {
	if cl == nil {
		return nil, ErrNoClient
	}

	return cl.Logical(), nil
}

func SimpleVaultLoad(addr, token, name string, to interface{}) error {
	cl, err := NewVaulter(addr, token)
	if err != nil {
		return err
	}

	return Vault{Client: cl}.Load(name, to)
}

// NewVaulter returns the Vaulter interface. If role_id is given it will call SetTokenAppRole
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

	return Vault{Client: cl.Logical()}, err
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
func (v Vault) Load(appName string, to interface{}) error {
	refVal, err := internal.GetReflectElem(to)
	if err != nil {
		return err
	}

	if err := v.LoadGeneric(to); err != nil {
		return err
	}

	return v.LoadReflect(getVaultSecretPath(appName), refVal)
}

// LoadGeneric loads generic(shared) secrets from Vault.
func (v Vault) LoadGeneric(to interface{}) error {
	refVal := reflect.Indirect(reflect.ValueOf(to))

	return v.LoadReflect(getVaultSecretPath(VaultSecretGenericPath), refVal)
}

// LoadReflect loads data from secret storage to refVal.
// It also works for inner structs.
func (v Vault) LoadReflect(appName string, refVal reflect.Value) error {
	secretMap, err := v.loadSecretData(appName)
	if err != nil {
		return err
	}

	t := refVal.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldVal, fieldTyp := refVal.Field(i), t.Field(i)

		if !fieldVal.CanSet() {
			if v.ErrOnUnsetable {
				return fmt.Errorf("cannot set field field: %q", fieldTyp.Name)
			}

			log.Warn().Str("field", fieldTyp.Name).Msg("cannot set field")

			continue
		}

		tags := internal.TagValue(fieldTyp, VaultSecretTag)
		if len(tags) == 0 {
			continue
		}

		tag := tags[0]

		if fieldTyp.Type.Kind() == reflect.Struct {
			if _, ok := SkipStructTypes[fieldTyp.Type]; ok {
				continue
			}

			if err := v.LoadReflect(appName+"/"+strings.ToLower(tag), fieldVal); err != nil {
				return fmt.Errorf("inner struct %q error: %w", fieldVal.Type().Name(), err)
			}

			continue
		}

		if err := setFieldValue(secretMap[tag], fieldTyp.Name, fieldVal); err != nil {
			return err
		}
	}

	return nil
}

func (v Vault) loadSecretData(appName string) (map[string]interface{}, error) {
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

func getVaultSecretPath(parts ...string) string {
	return path.Join(append([]string{VaultSecretBasePath}, parts...)...)
}

// setFieldValue sets value for reflect field.
func setFieldValue(val interface{}, fieldTypName string, fieldVal reflect.Value) error {
	if val == nil {
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		strVal = fmt.Sprint(val)
	}

	return internal.SetReflectValueString(fieldTypName, strVal, fieldVal)
}
