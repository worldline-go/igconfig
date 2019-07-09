package igconfig

import (
	"reflect"
	"strings"
	"strconv"
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// isTrue compares a string to determine the boolean value
func isTrue (s string) bool {
	return strings.Index(";TRUE;T;.T.;YES;Y;1;", ";" + strings.ToUpper(s) + ";") >= 0
}

// setValue sets a value in the config struct
func setValue(c interface{}, f reflect.StructField, v string) {
	k   := f.Name
	val := reflect.ValueOf(c).Elem().FieldByName(k)
	
	switch f.Type.Kind() {
		case reflect.Bool:
			val.SetBool(isTrue(v))
			
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				iglog.Warnf("LoadConfig: value for field %s not a valid integer", k)
			} else {
				val.SetInt(n)
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseUint(v, 0, 64)
			if err != nil {
				iglog.Warnf("LoadConfig: value for field %s not a valid unsigned integer", k)
			} else {
				val.SetUint(n)
			}

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				iglog.Warnf("LoadConfig: value for field %s not a valid float", k)
			} else {
				val.SetFloat(n)
			}

		case reflect.String:
			val.SetString(v)
			
		default:
			iglog.Warnf("LoadConfig: field %s unsupported type %s", f.Name, f.Type.Name())
	}

	if strings.ToUpper(k) == "LOGLEVEL" {
		switch strings.ToUpper(v) {
			case "DEBUGFULL"          : iglog.SetLevel(iglog.LogDebugFull)
			case "DEBUGMEDIUM"        : iglog.SetLevel(iglog.LogDebugMedium)
			case "DEBUGSMALL", "DEBUG": iglog.SetLevel(iglog.LogDebugSmall)
			case "INFO"               : iglog.SetLevel(iglog.LogInfo)
			case "WARN", "WARNING"    : iglog.SetLevel(iglog.LogWarn)
			case "ERROR"              : iglog.SetLevel(iglog.LogError)
			case "FATAL"              : iglog.SetLevel(iglog.LogFatal)
			case "PANIC"              : iglog.SetLevel(iglog.LogPanic)
			case "DEFAULT"            : iglog.SetLevel(iglog.LogDefault)
		}
	}
}
