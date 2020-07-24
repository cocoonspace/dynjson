package dynjson

import "reflect"

type sliceFormatter struct {
	t    reflect.Type
	elem formatter
}

func (f *sliceFormatter) typ() reflect.Type {
	return f.t
}

func (f *sliceFormatter) format(src reflect.Value) (reflect.Value, error) {
	dst := reflect.MakeSlice(f.t, src.Len(), src.Len())
	for i := 0; i < src.Len(); i++ {
		dv, err := f.elem.format(src.Index(i))
		if err != nil {
			return dv, err
		}
		dst.Index(i).Set(dv)
	}
	return dst, nil
}

type sliceBuilder struct {
	t    reflect.Type
	elem *structBuilder
}

func (b *sliceBuilder) build(fields []string) (formatter, error) {
	et, err := b.elem.build(fields)
	if err != nil {
		return nil, err
	}
	return &sliceFormatter{t: reflect.SliceOf(et.typ()), elem: et}, nil
}

func makeSliceBuilder(t reflect.Type) (*sliceBuilder, error) {
	elemBuilder, err := makeStructBuilder(t.Elem())
	if err != nil {
		return nil, err
	}
	return &sliceBuilder{t: t, elem: elemBuilder}, nil
}
