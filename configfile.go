package igconfig

import (
	"os"
	"bufio"
	"strings"
	"errors"
	"reflect"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// loadConfigFile loads config values from a file
func loadConfigFile(c interface{}, file string) error {
	iglog.DebugFullf("LoadConfig: loading config file '%s'", file)

	t := reflect.TypeOf(c)
	if t.Kind() != reflect.Struct {
		return errors.New("LoadConfig: input parameter is not a struct")
	}

	f, err := os.Open(file)
	if err != nil {
		iglog.Errorf("LoadConfig: could not open file '%s'", file)
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
				f := t.Field(i)
				if strings.ToUpper(f.Name) == k {
					setValue(c, f, v)
					break
				} else {
					nn := strings.Split(strings.ToUpper(f.Tag.Get("cfg")), ",")
					for _, n := range nn {
						if n == k {
							setValue(c, f, v)
							break
						}
					}
				}
			}
		}
	}

	return scanner.Err()
}
