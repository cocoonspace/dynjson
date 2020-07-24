package dynjson

import (
	"reflect"
)

type pointerFormatter struct {
	t    reflect.Type
	elem formatter
}

func (f *pointerFormatter) typ() reflect.Type {
	return f.t
}
func (f *pointerFormatter) format(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(f.t), nil
	}
	dst, err := f.elem.format(src.Elem())
	if err != nil {
		return dst, err
	}
	return dst.Addr(), nil
}

type pointerBuilder struct {
	t    reflect.Type
	elem *structBuilder
}

func (b *pointerBuilder) build(fields []string, prefix string) (formatter, error) {
	ef, err := b.elem.build(fields, prefix)
	if err != nil {
		return nil, err
	}
	return &pointerFormatter{t: reflect.PtrTo(ef.typ()), elem: ef}, nil
}

func makePointerBuilder(t reflect.Type) (*pointerBuilder, error) {
	eb, err := makeStructBuilder(t.Elem())
	if err != nil {
		return nil, err
	}
	return &pointerBuilder{t: t, elem: eb}, nil
}
