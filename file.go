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

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if strings.EqualFold(field.Name, k) {
				m.setValue(field.Name, v)
				break
			}
			nn := strings.Split(strings.ToUpper(field.Tag.Get("cfg")), ",")
			for _, n := range nn {
				if n == k {
					m.setValue(field.Name, v)
					break
				}
			}
		}
	}

	return scanner.Err()
}
