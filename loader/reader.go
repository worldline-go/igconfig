package loader

import (
	"bufio"
	"io"
	"os"
	"path"
	"strings"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/internal"
)

const CfgTag = internal.DefaultConfigTag

// Reader is intended to be a limited time option to read configuration from files.
// As such it is not included in default loaders list.
//
// Breaking changes from v1: config field name will be used as-is, without changing case.
//
// Deprecated: After Reader will be removed it's place will take Consul + Vault combination.
type Reader struct{}

// LoadEtc will load configuration file from /etc directory.
// File name is appName, so resulting path will be /etc/<appName>.
func (r Reader) LoadEtc(appName string, to interface{}) error {
	const etcPath = "/etc"

	return r.LoadFile(path.Join(etcPath, appName), to)
}

// LoadSecret will load Docker secret file.
// Full file path will be /run/secrets/<appName>.
func (r Reader) LoadSecret(appName string, to interface{}) error {
	const secretPath = "/run/secrets"

	return r.LoadFile(path.Join(secretPath, appName), to)
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

func (Reader) LoadReader(r io.Reader, to interface{}) error {
	refVal, err := internal.GetReflectElem(to)
	if err != nil {
		return err
	}

	t := refVal.Type()

	tagToFieldName := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tags := internal.TagValue(field, CfgTag)

		for _, n := range tags {
			tagToFieldName[n] = field.Name
		}
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()

		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, "//") || strings.HasPrefix(s, "#") {
			continue
		}

		i := strings.Index(s, "=")
		if i <= 0 {
			continue
		}

		k := strings.TrimSpace(s[:i])
		v := strings.TrimSpace(s[i+1:])
		fieldName, ok := tagToFieldName[k]

		if !ok {
			continue
		}

		if err := internal.SetStructFieldValue(fieldName, v, refVal); err != nil {
			return err
		}
	}

	return scanner.Err()
}
