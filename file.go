package igconfig

import (
	"bufio"
	"os"
	"reflect"
	"strings"
)

// loadFile loads config values from a fileName
func (m *localData) loadFile() error {
	v := reflect.ValueOf(m.userStruct)
	t := v.Elem().Type()

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
				m.fld = t.Field(i)
				if strings.EqualFold(m.fld.Name, k) {
					m.setValue(v)
					break
				}

				nn := strings.Split(strings.ToUpper(m.fld.Tag.Get("cfg")), ",")
				for _, n := range nn {
					if n == k {
						m.setValue(v)
						break
					}
				}
			}
		}
	}

	return scanner.Err()
}
