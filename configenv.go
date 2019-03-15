package igconfig

import (
	"reflect"
	"strings"
	"os"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// loadEnv loads config values from the environment
func loadEnv (c interface{}) {
	iglog.DebugFull("LoadConfig: loading environment variables")

	e := reflect.ValueOf(c).Elem()
	t := e.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !testEnv(c, f, f.Name) {
			if !testEnv(c, f, strings.ToLower(f.Name)) {
				if !testEnv(c, f, strings.ToUpper(f.Name)) {
					nn := strings.Split(strings.ToUpper(f.Tag.Get("env")), ",")
					for _, n := range nn {
						if testEnv(c, f, n) {
							break
						} else {
							if testEnv(c, f, strings.ToLower(n)) {
								break
							} else {
								if testEnv(c, f, strings.ToUpper(n)) {
									break
								}
							}
						}
					}
				}
			}
		}
	}
}

// testEnv tests for an environment variable, and if found sets the field's value
func testEnv(c interface{}, f reflect.StructField, n string) bool {
	v, ok := os.LookupEnv(n)
	if ok {
		setValue(c, f, v)
		return true
	}
	return false
}
