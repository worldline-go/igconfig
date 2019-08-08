package igconfig

import (
	"bufio"
	"errors"
	"os"
	"reflect"
	"strings"
)

// loadFile loads config values from a file
func (m *myConfig) loadFile() error {
	t := reflect.TypeOf(m.c)
	if t.Kind() != reflect.Struct {
		return errors.New("LoadConfig: input parameter is not a struct")
	}

	f, err := os.Open(m.file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()

		if i := strings.Index(s, "//"); i >= 0 {
			s = s[:i-1]
		}
		if i := strings.Index(s, "#"); i >= 0 {
			s = s[:i-1]
		}
		s = strings.TrimSpace(s)

		if i := strings.Index(s, "="); i > 0 {
			k := strings.ToUpper(strings.TrimSpace(s[:i-1]))
			v := strings.TrimSpace(s[i+1:])

			for i := 0; i < t.NumField(); i++ {
				m.f = t.Field(i)
				if strings.EqualFold(m.f.Name, k) {
					m.setValue(v)
					break
				}

				nn := strings.Split(strings.ToUpper(m.f.Tag.Get("cfg")), ",")
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
