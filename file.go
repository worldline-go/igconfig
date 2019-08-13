package igconfig

import (
	"bufio"
	"os"
	"strings"
)

// loadFile loads config values from a fileName
func (m *localData) loadFile() error {
	t := m.userStruct.Type()

	f, err := os.Open(m.fileName)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); e != nil && err == nil {
			err = e
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()

		s = strings.TrimSpace(s)

		if s == "" || strings.HasPrefix(s, "//") || strings.HasPrefix(s, "#") {
			continue
		}

		if i := strings.Index(s, "="); i > 0 {
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
	}

	return scanner.Err()
}
