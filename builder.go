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
		if t.Elem().Kind() != reflect.Struct {
			return makePrimitiveBuilder(t)
		}
		return makePointerBuilder(t)
	case reflect.Slice:
		if t.Elem().Kind() != reflect.Struct {
			return makePrimitiveBuilder(t)
		}
		return makeSliceBuilder(t)
	default:
		return makePrimitiveBuilder(t)
	}
}
