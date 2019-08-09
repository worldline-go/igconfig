package igconfig

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// loadFile loads config values from a file
func (m *myConfig) loadFile() error {
	const funcName = "loadFile"

	v := reflect.ValueOf(m.c)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("%s: input parameter not a pointer", funcName)
	}

	t := v.Elem().Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%s: input parameter not a struct", funcName)
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
