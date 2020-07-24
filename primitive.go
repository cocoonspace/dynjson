package dynjson

import "reflect"

type primitiveFormatter struct {
	t reflect.Type
}

func (f *primitiveFormatter) typ() reflect.Type {
	return f.t
}

func (f *primitiveFormatter) format(src reflect.Value) (reflect.Value, error) {
	return src, nil
}

type primitiveBuilder struct {
	t reflect.Type
}

func (b *primitiveBuilder) build(fields []string) (formatter, error) {
	return &primitiveFormatter{t: b.t}, nil
}

func makePrimitiveBuilder(t reflect.Type) (*primitiveBuilder, error) {
	return &primitiveBuilder{t: t}, nil
}
