package toolbox3d

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"time"
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
		case reflect.TypeOf(int64(0)):
			var defaultVal int64
			if defaultStr != "" {
				defaultVal, err = strconv.ParseInt(defaultStr, 10, 0)
				if err != nil {
					panic(err)
				}
			}
			f.Int64Var(val.FieldByIndex(field.Index).Addr().Interface().(*int64),
				flagName, defaultVal, usageStr)
		case reflect.TypeOf(uint(0)):
			var defaultVal uint
			if defaultStr != "" {
				x, err := strconv.ParseUint(defaultStr, 10, 64)
				if err != nil {
					panic(err)
				}
				defaultVal = uint(x)
			}
			f.UintVar(val.FieldByIndex(field.Index).Addr().Interface().(*uint),
				flagName, defaultVal, usageStr)
		case reflect.TypeOf(uint64(0)):
			var defaultVal uint64
			if defaultStr != "" {
				defaultVal, err = strconv.ParseUint(defaultStr, 10, 64)
				if err != nil {
					panic(err)
				}
			}
			f.Uint64Var(val.FieldByIndex(field.Index).Addr().Interface().(*uint64),
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
		case reflect.TypeOf(""):
			f.StringVar(val.FieldByIndex(field.Index).Addr().Interface().(*string),
				flagName, defaultStr, usageStr)
		case reflect.TypeOf(true):
			var defaultVal bool
			if defaultStr == "true" {
				defaultVal = true
			} else if defaultStr != "false" {
				panic(fmt.Sprintf("invalid boolean: %#v", defaultStr))
			}
			f.BoolVar(val.FieldByIndex(field.Index).Addr().Interface().(*bool),
				flagName, defaultVal, usageStr)
		case reflect.TypeOf(time.Duration(0)):
			var defaultVal time.Duration
			if defaultStr != "" {
				defaultVal, err = time.ParseDuration(defaultStr)
				if err != nil {
					panic(err)
				}
			}
			f.DurationVar(val.FieldByIndex(field.Index).Addr().Interface().(*time.Duration),
				flagName, defaultVal, usageStr)
		default:
			panic(fmt.Sprintf("unsupported argument type: %v", field.Type))
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
