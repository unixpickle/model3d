package toolbox3d

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

// AddFlags adds the numeric fields of a struct to a
// FlagSet.
//
// If f is nil, the command-line FlagSet is used.
//
// Fields should be annotated with `default:...` to
// indicate the default integer or floating point values.
//
// This may panic() if the default for a field is
// incorrectly formatted, or if a field is not supported.
func AddFlags(obj any, f *flag.FlagSet) {
	if f == nil {
		f = flag.CommandLine
	}
	val := reflect.ValueOf(obj).Elem()
	fields := reflect.VisibleFields(val.Type())
	for _, field := range fields {
		flagName := flagNameForField(field.Name)
		defaultStr := field.Tag.Get("default")
		usageStr := field.Tag.Get("usage")
		var err error
		switch field.Type {
		case reflect.TypeOf(int(0)):
			var defaultVal int
			if defaultStr != "" {
				defaultVal, err = strconv.Atoi(defaultStr)
				if err != nil {
					panic(err)
				}
			}
			f.IntVar(val.FieldByIndex(field.Index).Addr().Interface().(*int),
				flagName, defaultVal, usageStr)
		case reflect.TypeOf(float64(0)):
			var defaultVal float64
			if defaultStr != "" {
				defaultVal, err = strconv.ParseFloat(defaultStr, 64)
				if err != nil {
					panic(err)
				}
			}
			f.Float64Var(val.FieldByIndex(field.Index).Addr().Interface().(*float64),
				flagName, defaultVal, usageStr)
		default:
			panic(fmt.Sprintf("unsupported type: %v", field.Type))
		}
	}
}

func flagNameForField(field string) string {
	var result []rune
	for _, x := range field {
		if unicode.IsUpper(x) {
			if len(result) > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(x))
		} else {
			result = append(result, x)
		}
	}
	return string(result)
}
