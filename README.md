[![pipeline status](https://gitlab.test.igdcs.com/finops/utils/basics/iglog/badges/master/pipeline.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/iglog/commits/master)
[![coverage report](https://gitlab.test.igdcs.com/finops/utils/basics/iglog/badges/master/coverage.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/iglog/commits/master)

# igconfig package

igconfig package can be used to load configuration values from a configration file,
environment variables and/or command-line parameters.

## Requirements
This package requires the following packages:
- igstrings
- iglog

## Install

Clone this repo.
Install [GoLang] >= 1.12  (https://golang.org/doc/install).
```
cd <reponame>
go build *.go
```

## Unit tests
```
cd <reponame>
go test
```

## Code coverage report (Browser)
```
cd <reponame>
go test -coverprofile=cover.out
go tool cover -html=cover.out
```

## Description
There is only a single exported function:
```
func LoadConfig(c interface{}, file string, env bool, cmd bool, log bool) error
```
- c must be a pointer to struct or the function will fail with an error.
- file is the path to a configuration file (if an empty string, no file will be checked).
- env is a flag to indicate environment variables should be checked.
- cmd is a flag to indicate command-line parameters should be checked.
- log is a flag to indicate the loaded configuration structure should be logged.

### Config struct
The parameter c should be a pointer to a struct type.
All exported fields of this structure will be checked and filled if their name or key(s) in
the appropriate tag are identified within the configuration file, as an environment variable
or as a command-line parameter.
The field type is taken into consideration when processing parameters from a config file,
the environment and command-line parameters.
If the given value cannot be converted to the field's type, a warning will be logged,
and the value of the field remains as is.
Fields can be given a tag with identifier "default", where the value of the tag will then be
used as the default value for that field. If the default value cannot be converted to the
field's type, a warning will be logged.

### Config file
A config file should contain lines with "key=value" pairs.
Blank lines and lines that start with // or # are ignored.
The key is checked against the exported field names of the config struct and the field tag
identified by "cfg". The tag may contain a list of names separated by comma's.
Comparisons are done case-insensitive! 
If a key is matched, the corresponding field in the struct will be filled with the value
from the configuration file.
If a value for a config file is supplied but the file cannot be processed, LoadConfig will
return the file error, but will process environment variables and command-line parameters
if requested.

### Environment variables
For all exported fields from the config struct the name and the field tag identified by "env"
will be checked if a corresponding environment variable is present. The tag may contain
a list of names separated by comma's. The comparison is first done in a case-sensitive manner,
followed by all lower-case comparison, followed by all upper-case comparison.
Once a match is found the value from the corresponding environment variable is placed
in the struct field, and no further comparisons are done for that field.

### Command-line parameters
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
    Host: string `cfg:"hostname" env:"hostname", cmd:"h,host,hostname" default:"127.0.0.1"`
    Port: uint16 `cmd:"p,port" default:"8080"`
}
```
