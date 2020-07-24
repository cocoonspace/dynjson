package dynjson

import (
	"reflect"
)

type builder interface {
	build(fields []string) (formatter, error)
}

func makeBuilder(t reflect.Type) (builder, error) {
	switch t.Kind() {
	case reflect.Struct:
		return makeStructBuilder(t)
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Struct {
			return makePointerBuilder(t)
		} else {
			return makePrimitiveBuilder(t)
		}
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Struct {
			return makeSliceBuilder(t)
		} else {
			return makePrimitiveBuilder(t)
		}
	default:
		return makePrimitiveBuilder(t)
	}
}
