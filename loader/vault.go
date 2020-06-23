package loader

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
)

var VaultSecretTag = "secret"

// VaultSecretBasePath is the base path for secrets.
// This path MUST always end with trailing slash!
var VaultSecretBasePath = "secret/data/"

// VaultSecretGenericPath is path for storing generic secrets.
// Such generic secrets might be re-used by any number of applications.
// They must not be application specific!
var VaultSecretGenericPath = "generic"

// SkipStructTypes specifies types of structs that should be skipped from setting.
var SkipStructTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(time.Time{}): {},
}

type Vaulter interface {
	Read(path string) (*api.Secret, error)
}

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

func SimpleVaultLoad(addr, token, name string, to interface{}) error {
	cl, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return err
	}

	cl.SetToken(token)

	return Vault{Client: cl.Logical()}.Load(name, to)
}

// Load will load data from Vault to input struct 'to'.
// 'name' is base secret path, or just name of application.
// Path will be constructed as "${VaultSecretTag}/${name}".
// By default VaultSecretTag value is "secrets/data", which allows to load secrets from root.
func (v Vault) Load(name string, to interface{}) error {
	refVal := reflect.Indirect(reflect.ValueOf(to))

	return v.loadReflect(VaultSecretBasePath+name, refVal)
}

// LoadGeneric loads generic(shared) secrets from Vault.
func (v Vault) LoadGeneric(to interface{}) error {
	refVal := reflect.Indirect(reflect.ValueOf(to))

	return v.loadReflect(VaultSecretBasePath+VaultSecretGenericPath, refVal)
}

// loadReflect loads data from secret storage to refVal.
// It also works for inner structs.
func (v Vault) loadReflect(path string, refVal reflect.Value) error {
	pathSecret, err := v.Client.Read(path)
	if err != nil {
		return fmt.Errorf("vault request for path %q failed: %w", path, err)
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
		return fmt.Errorf("secret from path %q cannot be converted to map", path)
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

		tag := fieldTyp.Tag.Get(VaultSecretTag)
		switch tag {
		case "-":
			continue
		case "":
			tag = strings.ToLower(fieldTyp.Name)
		}

		if fieldTyp.Type.Kind() == reflect.Struct {
			if _, ok := SkipStructTypes[fieldTyp.Type]; ok {
				continue
			}

			if err := v.loadReflect(path+"/"+strings.ToLower(tag), fieldVal); err != nil {
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

// setFieldValue sets value for reflect field
func setFieldValue(val interface{}, fieldTypName string, fieldVal reflect.Value) error {
	if val == nil {
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		strVal = fmt.Sprint(val)
	}

	return setValue(fieldTypName, strVal, fieldVal)
}
