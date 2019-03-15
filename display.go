package igconfig

import (
	"reflect"
	"fmt"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// displayConfig displays the config struct fields with their assigned values
func displayConfig(c interface{}) {
	e := reflect.ValueOf(c).Elem()
	t := e.Type()

	m := []string {"Configuration loaded:"}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		k := f.Name
		val := e.FieldByName(f.Name)
		m = append(m, fmt.Sprintf("  %s=%v", k, val.Interface()))
	}

	iglog.Info(m...)
}
