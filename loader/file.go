package loader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/codec"
)

var FileTag = "cfg"

// ConfFileSuffixes is the ordered list of suffix for configuration file.
// It is not specific for type(.yml, .yaml, .json) because it is possible to change which loader will be used.
var ConfFileSuffixes = []string{".yml", ".yaml", ".json"}

// ErrNoDecoder is a serious error and not continue process.
var ErrNoDecoder = errors.New("decoder not found for this file type")

// ErrNoEnv use to skip reader.
var ErrNoEnv = errors.New("CONFIG_FILE not declared")

// ErrNoConfFile use to skip reader.
var ErrNoConfFile = errors.New("config file not found")

var _ Loader = File{}

// EnvConfigFile sets a name for environmental variable that can hold path for configuration file.
const EnvConfigFile = "CONFIG_FILE"

// File is intended to be a limited time option to read configuration from files.
// Set configuration path on CONFIG_FILE environment variable.
// '.yml|.yaml|.json' extensions supported.
//
// Breaking changes from v1: config field name will be used as-is, without changing case.
type File struct {
	// EtcPath for setting new path for etc, default is '/etc'.
	EtcPath string
	// NoFolderCheck doesn't try to check working directory and 'EtcPath'
	// with this formation '<appname>.[yml|yaml|json]'.
	NoFolderCheck bool
}

// LoadWithContext will try to load configuration file from two places: working directory(or file specified in env) and /etc.
// File in /etc will only be read if configuration file is missing in working directory.
//
// See DefaultDecoder for understanding of which decoder will used in this loader.
//
// Not existing configuration files are not treated as an error.
// If this behavior is required - use `Reader.Load*` methods directly.
func (l File) LoadWithContext(_ context.Context, appName string, to interface{}) error {
	// check ENV file
	err := l.LoadEnv(to)
	if !errors.Is(err, ErrNoEnv) {
		return err
	}

	if l.NoFolderCheck {
		return nil
	}

	// check working directory
	err = l.LoadWorkDir(appName, to)
	if !errors.Is(err, ErrNoConfFile) {
		return err
	}

	// check etc directory
	err = l.LoadEtc(appName, to)
	if !errors.Is(err, ErrNoConfFile) {
		return err
	}

	workingDir, _ := os.Getwd()

	return fmt.Errorf("%w: %s not found in %s and %s", ErrNoConfFile, appName, workingDir, l.EtcPath)
}

// Load is just same as LoadWithContext without context.
func (l File) Load(appName string, to interface{}) error {
	return l.LoadWithContext(context.TODO(), appName, to)
}

// LoadWorkDir will load configuration from current working directory.
//
// If EnvConfigFile environment variable is specified - it will have priority over current working directory.
// EnvConfigFile value will be expanded with environment before use.
func (l File) LoadWorkDir(appName string, to interface{}) error {
	return l.LoadFileSuffix(cleanName(appName), to)
}

// LoadEtc will load configuration file from /etc directory.
// File name is appName, so resulting path will be /etc/<appName>.
func (l *File) LoadEtc(appName string, to interface{}) error {
	appName = cleanName(appName)

	if l.EtcPath == "" {
		l.EtcPath = "/etc"
	}

	filePath := path.Join(l.EtcPath, appName)

	return l.LoadFileSuffix(filePath, to)
}

func (l File) LoadFileSuffix(filePath string, to interface{}) error {
	for _, s := range ConfFileSuffixes {
		if _, err := os.Stat(filePath + s); !os.IsNotExist(err) {
			// suffix = s
			return l.LoadFile(filePath+s, to)
		}
	}

	return ErrNoConfFile
}

func (l File) LoadEnv(to interface{}) error {
	if envFile := os.Getenv(EnvConfigFile); envFile != "" {
		return l.LoadFile(envFile, to)
	}

	return ErrNoEnv
}

// LoadFile loads config values from a fileName.
func (l File) LoadFile(fileName string, to interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("file loader: %w", err)
	}
	defer file.Close()

	return l.loadReader(file, to, filepath.Ext(fileName))
}

// loadReader automatically choice decoder with config file suffix.
func (l File) loadReader(reader io.Reader, to interface{}, configType string) error {
	var decoder codec.Decoder

	switch configType {
	case ".yaml", ".yml":
		decoder = codec.YAML{}
	case ".json":
		decoder = codec.JSON{}
	default:
		return fmt.Errorf("%w: %s", ErrNoDecoder, configType)
	}

	if err := codec.LoadReaderWithDecoder(reader, to, decoder, FileTag); err != nil {
		return fmt.Errorf("File.loadReader error: %w", err)
	}

	return nil
}

func cleanName(str string) string {
	str = strings.TrimSpace(str)
	str = strings.Trim(str, "/\\")

	idx := strings.LastIndex(str, "/")
	if idx == -1 {
		return str
	}

	return str[idx+1:]
}
