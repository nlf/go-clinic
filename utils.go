package clinic

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pborman/getopt"
)

func fieldLong(field reflect.StructField) string {
	if long, ok := field.Tag.Lookup("long"); ok {
		return long
	}

	return strings.ToLower(field.Name)
}

func fieldShort(field reflect.StructField) rune {
	if short, ok := field.Tag.Lookup("short"); ok {
		return rune(short[0])
	}

	return 0
}

func fieldUsage(field reflect.StructField) string {
	if usage, ok := field.Tag.Lookup("usage"); ok {
		return usage
	}

	return field.Name
}

func promptAllowed(field reflect.StructField) bool {
	if prompt, ok := field.Tag.Lookup("prompt"); ok {
		return strings.ToLower(prompt) == "yes"
	}

	return false
}

func parseFields(config interface{}) []field {
	if config == nil {
		return []field{}
	}

	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	fields := make([]field, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)

		fields[i] = field{
			value:  fieldVal,
			typ:    fieldTyp,
			iface:  fieldVal.Addr().Interface(),
			short:  fieldShort(fieldTyp),
			long:   fieldLong(fieldTyp),
			usage:  fieldUsage(fieldTyp),
			prompt: promptAllowed(fieldTyp),
		}
	}

	return fields
}

func addToFlags(fields []field, flags *getopt.Set) {
	for _, f := range fields {
		switch f.iface.(type) {
		case *bool:
			flags.BoolVarLong(f.iface.(*bool), f.long, f.short, f.usage)
		case *string:
			flags.StringVarLong(f.iface.(*string), f.long, f.short, f.usage)
		case *int:
			flags.IntVarLong(f.iface.(*int), f.long, f.short, f.usage)
		case *uint:
			flags.UintVarLong(f.iface.(*uint), f.long, f.short, f.usage)
		case *[]string:
			flags.ListVarLong(f.iface.(*[]string), f.long, f.short, f.usage)
		}
	}
}

func acceptsArgs(fn interface{}) bool {
	if fn == nil {
		return false
	}

	typ := reflect.TypeOf(fn)
	for i := 0; i < typ.NumIn(); i++ {
		if typ.In(i) == typStringSlice {
			return true
		}
	}

	return false
}

func getOptions(fields []field) [][]string {
	options := [][]string{}
	for _, f := range fields {
		var flag string
		if f.short != 0 {
			flag = fmt.Sprintf("-%s, --%s", string(f.short), f.long)
		} else {
			flag = fmt.Sprintf("--%s", f.long)
		}

		val := f.value.Interface()
		switch val.(type) {
		case string:
			flag += fmt.Sprintf(" [%q]", val)
		case []string:
			flag += " [\"" + strings.Join(val.([]string), "\", \"") + "\"]"
		default:
			flag += fmt.Sprintf(" [%v]", val)
		}

		options = append(options, []string{flag, f.usage})
	}

	return options
}

func interfaceToStringSlice(in interface{}) []string {
	inslice := in.([]interface{})
	ret := make([]string, len(inslice))
	for i := range inslice {
		ret[i] = inslice[i].(string)
	}

	return ret
}

func prompt(p string) string {
	var prefix string
	if stdoutColor {
		prefix = fmt.Sprintf("%s[?]%s", blue, reset)
	} else {
		prefix = "[?]"
	}
	var res string
	fmt.Printf("%s %s: ", prefix, p)
	fmt.Scanln(&res)
	return res
}

func promptBool(p string) bool {
	res := strings.ToLower(prompt(fmt.Sprintf("%s (yes/no)", p)))
	if res == "yes" || res == "y" || res == "true" || res == "t" {
		return true
	} else if res == "no" || res == "n" || res == "false" || res == "f" {
		return false
	} else {
		return promptBool(p)
	}
}

func promptInt(p string) int {
	res := prompt(fmt.Sprintf("%s (number)", p))
	i, err := strconv.ParseInt(res, 10, 0)
	if err != nil {
		return promptInt(p)
	}

	return int(i)
}

func promptUint(p string) uint {
	res := prompt(fmt.Sprintf("%s (number)", p))
	i, err := strconv.ParseUint(res, 10, 0)
	if err != nil {
		return promptUint(p)
	}

	return uint(i)
}

func promptStringSlice(p string) []string {
	var res []string
	for {
		r := prompt(fmt.Sprintf("%s (enter a blank line when finished)", p))
		if r == "" {
			break
		}
		res = append(res, r)
	}

	return res
}

func promptForMissing(fields []field, segment map[interface{}]interface{}, flags *getopt.Set) {
	for _, f := range fields {
		opt := flags.Lookup(f.long)
		if opt.Seen() {
			continue
		}

		if _, ok := segment[f.long]; ok {
			continue
		}

		if !f.prompt {
			continue
		}

		switch f.iface.(type) {
		case *bool:
			*f.iface.(*bool) = promptBool(f.usage)
		case *string:
			*f.iface.(*string) = prompt(f.usage)
		case *int:
			*f.iface.(*int) = promptInt(f.usage)
		case *uint:
			*f.iface.(*uint) = promptUint(f.usage)
		case *[]string:
			*f.iface.(*[]string) = promptStringSlice(f.usage)
		}
	}
}
