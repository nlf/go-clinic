package clinic

import (
	"reflect"
)

type field struct {
	value  reflect.Value
	typ    reflect.StructField
	iface  interface{}
	short  rune
	long   string
	usage  string
	prompt bool
}

func (f *field) set(val interface{}) {
	switch f.iface.(type) {
	case *bool:
		*f.iface.(*bool) = val.(bool)
	case *string:
		*f.iface.(*string) = val.(string)
	case *int:
		*f.iface.(*int) = val.(int)
	case *uint:
		*f.iface.(*uint) = val.(uint)
	case *[]string:
		*f.iface.(*[]string) = interfaceToStringSlice(val)
	}
}
