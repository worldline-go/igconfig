package loader

import (
	"errors"
	"io"
	"os"
	"path"
)

// ConfFileSuffix is the suffix for configuration file.
// It is not specific for type(.yaml, .json, etc) because it is possible to change which loader will be used.
var ConfFileSuffix = ".conf"

var _ Loader = Reader{}

// EnvConfigFile sets a name for environmental variable that can hold path for configuration file.
const EnvConfigFile = "CONFIG_FILE"

// Reader is intended to be a limited time option to read configuration from files.
// As such it is not included in default loaders list.
//
// Breaking changes from v1: config field name will be used as-is, without changing case.
//
// Config file name will be generated as "appName + ConfFileSuffix".
type Reader struct{}

// Load will try to load configuration file from two places: working directory(or file specified in env) and /etc.
// File in /etc will only be read if configuration file is missing in working directory.
//
// See DefaultDecoder for understanding of which decoder will used in this loader.
//
// Not existing configuration files are not treated as an error.
// If this behavior is required - use `Reader.Load*` methods directly.
func (r Reader) Load(appName string, to interface{}) error {
	err := CheckNotExistError(r.LoadWorkDir(appName, to))
	if err == nil {
		return nil
	}

	return CheckNotExistError(r.LoadEtc(appName, to))
}

// LoadWorkDir will load configuration from current working directory.
//
// If EnvConfigFile environment variable is specified - it will have priority over current working directory.
// EnvConfigFile value will be expanded with environment before use.
func (r Reader) LoadWorkDir(appName string, to interface{}) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath := path.Join(wd, appName+ConfFileSuffix)

	if envFile := os.Getenv(EnvConfigFile); envFile != "" {
		filePath = os.ExpandEnv(envFile)
	}

	return r.LoadFile(filePath, to)
}

// LoadEtc will load configuration file from /etc directory.
// File name is appName, so resulting path will be /etc/<appName>.
func (r Reader) LoadEtc(appName string, to interface{}) error {
	const etcPath = "/etc"

	return r.LoadFile(path.Join(etcPath, appName+ConfFileSuffix), to)
}

// LoadFile loads config values from a fileName.
func (r Reader) LoadFile(fileName string, to interface{}) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	return r.LoadReader(f, to)
}

// LoadReader is alias to Reader.LoadReaderWithDecoder(reader, to, DefaultDecoder)
func (r Reader) LoadReader(reader io.Reader, to interface{}) error {
	return r.LoadReaderWithDecoder(reader, to, DefaultDecoder)
}

// LoadReaderWithDecoder will decode input in `r` into `to` by using `decoder`.
func (Reader) LoadReaderWithDecoder(r io.Reader, to interface{}, decoder Decoder) error {
	return decoder(r, to)
}

// CheckNotExistError will return specified error if it is not os.ErrNotExists.
//
// Otherwise it will return nil.
func CheckNotExistError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
