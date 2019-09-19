package igconfig

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// loadFile loads config values from a fileName
func (m *localData) loadFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.loadReader(f)
}

func (m *localData) loadReader(r io.Reader) error {
	t := m.userStruct.Type()

	tagToFieldName := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tags := field.Tag.Get("cfg")
		if tags == "" {
			tagToFieldName[strings.ToUpper(field.Name)] = field.Name
			continue
		}
		nn := strings.Split(strings.ToUpper(tags), ",")
		for _, n := range nn {
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
		k := strings.ToUpper(strings.TrimSpace(s[:i]))
		v := strings.TrimSpace(s[i+1:])
		fieldName, ok := tagToFieldName[k]

		if !ok {
			continue
		}
		m.setValue(fieldName, v)
	}
	return scanner.Err()
}
