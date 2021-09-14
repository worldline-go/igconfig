[![pipeline status](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/badges/master/pipeline.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/commits/master)
[![coverage report](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/badges/master/coverage.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/commits/master)
[![Quality Gate Status](https://am2vm2329.test.igdcs.com/api/project_badges/measure?project=utils%2Fbasics%2Figconfig&metric=alert_status)](https://am2vm2329.test.igdcs.com/dashboard?id=utils%2Fbasics%2Figconfig)

# igconfig package

igconfig package can be used to load configuration values from a configuration file,
environment variables, Consul, Vault and/or command-line parameters.

## Requirements
This package does not require any external packages.

## Install
Add this package to `go.mod`:

```go
require (
 gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2 latest
)
```

<details><summary>Tests</summary>

## Unit tests
```sh
go test ./...
```

## Code coverage report (Browser)
```sh
mkdir _out
go test -cover -coverprofile cover.out -outputdir ./_out/ ./...
go tool cover -html=./_out/cover.out
# Export HTML
# go tool cover -html=./_out/cover.out -o ./_out/coverage.html
```

</details>

## Description
There is only a single exported function:
```
func LoadConfig(appName string, config interface{}) error
```
or if specific loaders needed:
```
func LoadWithLoaders(appName string, configStruct interface{}, loaders ...loader.Loader) error
```

- `appName` is name of application. It is used in Consul and Vault to find proper path for variables.
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
This will be shown later.  

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

### Default
This loader uses `default` tag to get value for fields.

### Consul
Loads configuration from Consul and uses YAML to decode data from Consul to a struct.

While it is possible to change decoder from YAML to JSON for example it is not recommended 
if there are no objective reasons to do so. YAML is superior to JSON in terms of readability 
while providing as much ability to write configurations.

For better configurability configuration struct might include `yaml` tag for fields to 
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
field1: 50
str:
    inner: "test string"
```

### Vault
Vault loads data from map, and while Vault provides ability to store secrets as "code" 
this library is not able to decode "code" secrets.

### File
YAML and JSON files supported, and file path should be located on __CONFIG_FILE__ env variable.  
If that environment variable not found, file loader check working directory and `/etc` path
with this formation `<appName>.[yml|yaml|json]`.  
The appName used as the file name is not the full name, only the part after the last slash.
So if your app name is `transactions/consumers/internal/apm`,
the loader will try to load a file with the name `apm`.

The key is checked against the exported field names of the config struct and the field tag
identified by `cfg`.  
If a key is matched, the corresponding field in the struct will be filled with the value
from the configuration file.

### Environment variables
For all exported fields from the config struct the name and the field tag identified by "env"
will be checked if a corresponding environment variable is present. The tag may contain
a list of names separated by comma's. The comparison is upper-case, 
even if tag specifies lower- or mixed-case.
Once a match is found the value from the corresponding environment variable is placed
in the struct field, and no further comparisons will be done for that field.

### Flags (command-line parameters)
For all exported fields from the config struct the tag of the field identified by "cmd"
will be checked if a corresponding command-line parameter is present. The tag may contain
a list of names separated by comma's. The comparison is always done in a case-sensitive manner.
Once a match is found the value from the corresponding command-line parameter is placed
in the struct field, and no further comparisons are done for that field.
For boolean struct fields the command-line parameter is checked as a flag.
For all other field types the command-line parameter should have a compatible value.
Parameters can be supplied on the command-line as described in the standard Go package "flag".

### Example config struct
```
type MyConfig struct {
    Host string `cfg:"hostname" env:"hostname", cmd:"h,host,hostname" default:"127.0.0.1"`
    Port uint16 `cfg:"port" default:"8080"` // Will also define flags and will search in env based on 'cmd' tag
}
```
