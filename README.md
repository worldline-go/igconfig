# igconfig

[![Codecov](https://img.shields.io/codecov/c/github/worldline-go/igconfig?logo=codecov&style=flat-square)](https://app.codecov.io/gh/worldline-go/igconfig)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/worldline-go/igconfig/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/worldline-go/igconfig/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/worldline-go/igconfig.svg)](https://pkg.go.dev/github.com/worldline-go/igconfig)

This package can be used to load configuration values from a configuration file,
environment variables, Consul, Vault and/or command-line parameters.

## Install

```sh
go get github.com/worldline-go/igconfig
```

## Example

**cfg** and **secret** tag values are case insensitive and weakly dash/underscore so **Network-Name**, **network_name**, **NeTWoK-NaMe** and **NeTWoKNaMe** are same in both tag and configs.

**NOTE** if **secret** tag not found it will check **cfg** tag after that it will check variable's name.

```go
type Config struct {
	NetworkName string `cfg:"networkName" env:"NETWORK_NAME" secret:"networkName"`
	// application specific vault
	DBSchema     string `cfg:"dbSchema"     env:"SCHEMA"       secret:"dbSchema,loggable" default:"transaction"`
	DBDataSource string `cfg:"dbDataSource" env:"DBDATASOURCE" loggable:"false"`
	DBType       string `cfg:"dbType"       env:"DBTYPE"       default:"pgx"`

	CustomConfig map[string]interface{} `cfg:"customConfig" secret:"customConfig,loggable"`
}

// ---

var cfg Config

if err := igconfig.LoadConfig("myappname", &cfg); err != nil {
    log.Fatal().Err(err).Msg("unable to load configuration settings.")
}
```

Also check example:  
[Examples section](#examples)  
[Example usage \_example/readFromAll/main.go](_example/readFromAll/main.go)

## Description

There is only a single exported function:

```go
func LoadConfig(appName string, config interface{}) error
```

or if specific loaders needed:

```go
func LoadWithLoaders(appName string, configStruct interface{}, loaders ...loader.Loader) error
```

There are also context accepted functions `LoadConfigWithContext`, `LoadWithLoadersWithContext`.

- `appName` is name of application. It is used in Consul and Vault to find proper path for variables also file name for file loader.
- `config` must be a pointer to struct. Otherwise, the function will fail with an error.
- `loaders` is list of Loaders to use.

### Config struct

All exported fields of this structure will be checked and filled based on their tags or field names.

The field type is taken into consideration when processing parameters.

If the given value cannot be converted to the field's type, an error will be returned.
Failing to provide proper value for a type is human-made error, which means that something is not right.

Fields can be given a tag with identifier "default", where the value of the tag will then be
used as the default value for that field. If the default value cannot be converted to the
field's type, an error will be returned.

Config struct can have inner structs as a fields, but not all Loaders might support them.
For example Consul and Vault support inner structs, while Env and Flags don't.

### Tags

Config structs can have special tags for fine-grained field name configuration.

There are no required tags, but setting them can improve readability and understanding.

#### cfg

`cfg` tag is fallback tag when no Loader-specific tag can be found.
As such defining only this tag can be enough for most situations.

#### env

`env` tag specifies a name of environmental variable to get value from.

#### cmd

`cmd` tag is used to set flag names for fields.

#### secret

`secret` tag specifies name of field in Vault that should be used to fill the field.  
If not exist it use as struct's field name.

#### default

`default` is special tag.

Unlike other tags it does not point to a place from which value should be taken,
but instead it itself holds value.

`default:"data"` will mean that value of string field that has this tag will be `data`.

This tag is optional

## Loaders

Loaders are actual specification on how fields should be filled.

`igconfig` provides a simple interface for creating new loaders.

Below is a sorted list of currently provided loaders that are included by default(if not stated otherwise)

Change order of loaders and configurations:

```go
// this is the default loaders; change configurations and order or eliminate some loaders
loaders := []loader.Loader{
	&loader.Default{},
	&loader.Consul{},
	&loader.Vault{},
	&loader.File{},
	&loader.Env{},
	&loader.Flags{},
}

// read configurations with custom loaders
if err := igconfig.LoadWithLoaders("test", &conf, loaders...); err != nil {
    log.Fatal().Err(err).Msg("unable to load configuration settings.")
}
```

### Default

This loader uses `default` tag to get value for fields.

### Consul

Loads configuration from Consul and uses map decoder with `cfg` tag to decode data from Consul to a struct.

If not give `CONSUL_HTTP_ADDR` as environment variable, this config will skip!

For connection to Consul server you need to set some of environment variables.

| Envrionment variable      | Meaning                                                                        |
| ------------------------- | ------------------------------------------------------------------------------ |
| CONSUL_HTTP_ADDR          | Ex: `consul:8500`, sets the HTTP address                                       |
| CONSUL_HTTP_TOKEN_FILE    | sets the HTTP token file                                                       |
| CONSUL_HTTP_TOKEN         | sets the HTTP token                                                            |
| CONSUL_HTTP_AUTH          | Ex: `username:password`, sets the HTTP authentication header                   |
| CONSUL_HTTP_SSL           | Ex: `true`, sets whether or not to use HTTPS                                   |
| CONSUL_TLS_SERVER_NAME    | sets the server name to use as the SNI host when connecting via TLS            |
| CONSUL_CACERT             | sets the CA file to use for talking to Consul over TLS                         |
| CONSUL_CAPATH             | sets the path to a directory of CA certs to use for talking to Consul over TLS |
| CONSUL_CLIENT_CERT        | sets the client cert file to use for talking to Consul over TLS                |
| CONSUL_CLIENT_KEY         | sets the client key file to use for talking to Consul over TLS.                |
| CONSUL_HTTP_SSL_VERIFY    | Ex: `false`, sets whether or not to disable certificate checking               |
| CONSUL_NAMESPACE          | sets the HTTP Namespace to be used by default. This can still be overridden    |
| CONSUL_CONFIG_PATH_PREFIX | sets the path prefix to be used by default.                                    |

While it is possible to change decoder from YAML to JSON for example it is not recommended
if there are no objective reasons to do so. YAML is superior to JSON in terms of readability
while providing as much ability to write configurations.

For better configurability configuration struct might include `cfg` tag for fields to
specify a proper name to bind from Consul, if this tag is skipper - lowercase field name will be used to bind.

For example:

```go
type Config struct {
    Field1 int `cfg:"field"`
    Str struct {
        Inner string
    }
}
```

will match this YAML

```yaml
field: 50
str:
  inner: "test string"
```

### Vault

Loads configuration from Vault and uses MapDecoder to decode data from Vault to a struct.

First Vault loads in `finops/data/generic` path and after that process application's configuration in `finops/data/<appname>` path.

`generic` path can have inner path, vault loader combine them.

To use additional path to load with appname set `AdditionalPaths` value in the loader.  
Default value is `[{Map: "", Name: "generic"}]`, Map is a wrapper for read value in key-value format, generic doesn't have map value so it will apply what read and append directly in our config.

To read more than one path, just append your path to the loader.VaultSecretAdditionalPaths slice.

Use `Map` value to wrap readed data with a key and `Name` is a path of configuration.

```go
loader.VaultSecretAdditionalPaths = append(
    loader.VaultSecretAdditionalPaths,
    loader.AdditionalPath{Map: "loadtest", Name: "loadtest"},
)
```

If not given any of `VAULT_ADDR`, `VAULT_AGENT_ADDR` or `CONSUL_HTTP_ADDR` as environment variable, this config will skip!

If `CONSUL_HTTP_ADDR` exists, it uses Consul to get vault address.

| Envrionment Variable      | Meaning                                                                                                                                                                                        |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CONSUL_HTTP_ADDR          | get VAULT_ADDR from this consul server with vault service tag name.                                                                                                                            |
| VAULT_CONSUL_ADDR_DISABLE | disable to get VAULT_ADDR from this consul server.                                                                                                                                             |
| VAULT_ADDR                | the address of the Vault server. This should be a complete URL such as "http://vault.example.com". If need a custom SSL cert or enable insecure mode, you need to specify a custom HttpClient. |
| VAULT_AGENT_ADDR          | the address of the local Vault agent. This should be a complete URL such as "http://vault.example.com".                                                                                        |
| VAULT_MAX_RETRIES         | controls the maximum number of times to retry when a 5xx error occurs. Set to 0 to disable retrying. Defaults to 2 (for a total of three tries).                                               |
| VAULT_RATE_LIMIT          | EX: `rateFloat:brustInt`                                                                                                                                                                       |
| VAULT_CLIENT_TIMEOUT      | seconds                                                                                                                                                                                        |
| VAULT_SRV_LOOKUP          | enables the client to lookup the host through DNS SRV lookup                                                                                                                                   |
| VAULT_CACERT              | TLS                                                                                                                                                                                            |
| VAULT_CAPATH              | TLS                                                                                                                                                                                            |
| VAULT_CLIENT_CERT         | TLS                                                                                                                                                                                            |
| VAULT_CAPATH              | TLS                                                                                                                                                                                            |
| VAULT_CLIENT_CERT         | TLS                                                                                                                                                                                            |
| VAULT_CLIENT_KEY          | TLS                                                                                                                                                                                            |
| VAULT_TLS_SERVER_NAME     | TLS                                                                                                                                                                                            |
| VAULT_SKIP_VERIFY         | TLS                                                                                                                                                                                            |
| VAULT_APPROLE_BASE_PATH   | set login path to be used by default.                                                                                                                                                          |
| VAULT_SECRET_BASE_PATH    | set secret base path to be used by default.                                                                                                                                                    |

For authentication, you should set `VAULT_ROLE_ID` and `VAULT_ROLE_SECRET` environment variables.

### File

TOML, YAML and JSON files supported, and file path should be located on **CONFIG_FILE** env variable.  
If that environment variable not found, file loader check working directory and `/etc` path
with this formation `<appName>.[toml|yml|yaml|json]` (if there is more than `appName` with different suffixes, order is `toml > yml > yaml > json`).  
The appName used as the file name is not the full name, only the part after the last slash.
So if your app name is `transactions/consumers/internal/apm/`,
the loader will try to load a file with the name `apm`.

The key is checked against the exported field names of the config struct and the field tag
identified by `cfg`.  
If a key is matched, the corresponding field in the struct will be filled with the value
from the configuration file.

**NOTE:** if `cfg` tag not exists, it is still read values in file and match struct's field name!  
Don't want to read a value just delete it in your config file or add `cfg:"-"`.

FileLoader editable, you can add your own decoder or new file format or order of file suffixes.

### Environment variables

For all exported fields from the config struct the name and the field tag identified by "env"
will be checked if a corresponding environment variable is present. The tag may contain
a list of names separated by comma's. The comparison is upper-case,
even if tag specifies lower- or mixed-case. Lower-case environment variables are ignored.

Once a match is found the value from the corresponding environment variable is placed
in the struct field, and no further comparisons will be done for that field.

Set value in inner struct:

```go
type Config struct {
    Inner Inner
}

type Inner struct {
	GetENV       string `env:"TEST_ENV"` // env:"test_EnV" same as TEST_ENV
	// GetENV       string `cfg:"TEST_ENV"` // cfg tag also usable
}
```

To set `GetENV` value use `INNER_TEST_ENV` environment value. Or you can change `INNER` name with env tag.

```go
type Config struct {
    Inner Inner `env:"IN"`
}
```

Now value use `IN_TEST_ENV`

**NOTE:** if `env` tag not exists, it will check `cfg` tag and if both not exists, it will check struct's field name as uppercase.

### Flags (command-line parameters)

For all exported fields from the config struct the tag of the field identified by "cmd"
will be checked if a corresponding command-line parameter is present. The tag may contain
a list of names separated by comma's. The comparison is always done in a case-sensitive manner.
Once a match is found the value from the corresponding command-line parameter is placed
in the struct field, and no further comparisons are done for that field.
For boolean struct fields the command-line parameter is checked as a flag.
For all other field types the command-line parameter should have a compatible value.
Parameters can be supplied on the command-line as described in the standard Go package "flag".

## Log with context

Set a new zerolog logger and attach to the context, igconfig will use that context's logger.

```go
// set new schemas for log
logConfig := log.With().Str("component", "config").Logger()
// replace context.Background() with own context
logCtx := logConfig.WithContext(context.Background())

// call igconfig with context
if err := igconfig.LoadConfigWithContext(logCtx, "test", &conf); err != nil {
    log.Ctx(logCtx).Fatal().Err(err).Msg("unable to load configuration settings.")
}
```

## Print configuration

`secret` tag is disabled to print but if you want to print it add aditional option to secret called `loggable`.

```go
Info string     `cfg:"info" secret:"info,loggable"` // Set secret's loggable option
```

Or you can set general loggable to manage it all tags in releated field.

```go
User string     `cfg:"user" secret:"user" loggable:"true"` // Set general loggable
```

```go
conf := config.AppConfig{} // set up config value somehow
// log is zerolog/log package
log.Info().
    EmbedObject(Printer{Value: conf}).
    Msg("loaded config")
```

## Examples

<details><summary>Example usage of File</summary>

```sh
(
export CONFIG_FILE=_example/readFromAll/dataFile/test.yml
go run _example/readFromAll/main.go
)
```

</details>

<details><summary>Example usage of Vault server</summary>

Set Vault or Consul server address with releated environment variables.

Run a development vault

```sh
docker run -it --rm --cap-add=IPC_LOCK --name=dev-vault -p 8200:8200 vault
```

After that connect this vault with vault CLI app.

```sh
# export address for http
export VAULT_ADDR="http://127.0.0.1:8200"
# login with root token (appears in docker output)
vault login <token>
# unseal it
vault operator unseal <unsealkey>
# create kv secret engine
vault secrets enable -path=finops -version=2 kv
# create policy to read
{
cat <<EOF
path "finops/*" {
  capabilities = ["read", "list"]
}
path "finops/data/generic/super-secret" {
  capabilities = ["deny"]
}
EOF
} | vault policy write finops-read -

# create a approle with policy and enable connection without secret_id
vault auth enable approle
vault write auth/approle/role/my-role bind_secret_id=false secret_id_bound_cidrs="127.0.0.0/8,172.17.0.0/16" policies="default","finops-read"
# learn role-id
ROLE_ID=$(vault read -field=role_id auth/approle/role/my-role/role-id)

# fill some data
vault kv put finops/generic/keycloack @_example/readFromAll/dataVault/generic_keycloack.json
vault kv put finops/generic/super-secret @_example/readFromAll/dataVault/generic_supersecret.json
vault kv put finops/test @_example/readFromAll/dataVault/test.json
vault kv put finops/loadtest @_example/readFromAll/dataVault/loadtest.json
```

After that add our data in your `finops` kv section. Under usually should be a `generic` section and you should add keycloack and migration in there. also add your application name data in `finops`.

```sh
(
export VAULT_ADDR="http://localhost:8200"
export VAULT_ROLE_ID=${ROLE_ID}
# export CONSUL_HTTP_ADDR="am2vm2042.test.igdcs.com:8500"
export MIGRATIONS_TEST_ENV="testing_testing_1234"
go run _example/readFromAll/main.go
)
```

</details>

<details><summary>Example usage of Consul server</summary>

Start consul agent with dev mode

```sh
docker run -it --rm --name=dev-consul --net=host consul:1.10.4
```

Go to `localhost:8500` webui and add key values but for our tool folder should be `finops`.

It could be `yaml` or `json` format or you can handle by `codec.Decoder` interface.

Test it

```sh
export CONSUL_HTTP_ADDR="localhost:8500"
go run _example/readFromAll/main.go
```

</details>

<details><summary>Example usage of Consul dynamic listen</summary>

To listen a key, our function is easily usable.  
While listening a key, you can restart consul server or close or not even started yet or delete key, it is totally safe.

Check example to get information:

```sh
export CONSUL_HTTP_ADDR="localhost:8500"
go run _example/dynamicConsul/main.go
```

</details>

## Development

<details><summary>Package Test</summary>

## Unit tests

```sh
go test --race -cover ./...
```

## Code coverage report (Browser)

```sh
mkdir _out
go test -cover -coverprofile cover.out -outputdir ./_out/ ./...
# Auto open html result
go tool cover -html=./_out/cover.out
# Export HTML
# go tool cover -html=./_out/cover.out -o ./_out/coverage.html
```

</details>

<details><summary>JSON|YAML</summary>

Use `yq` tool for translate json to yaml or yaml to json.

```sh
# yaml to json
cat _example/readFromAll/dataFile/test.yml | yq
```

```sh
# json to yaml
cat _example/readFromAll/dataVault/generic_keycloack.json | yq -y
```

</details>
