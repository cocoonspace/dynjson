package dynjson

import (
	"fmt"
	"reflect"
	"strings"
)

type mapping struct {
	src    reflect.StructField
	dst    reflect.StructField
	format formatter
}

type structFormatter struct {
	t        reflect.Type
	mappings map[string]mapping
}

func (f *structFormatter) typ() reflect.Type {
	return f.t
}

func (f *structFormatter) format(src reflect.Value) (reflect.Value, error) {
	pdst := reflect.New(f.t)
	dst := pdst.Elem()
	for key := range f.mappings {
		sv := src.FieldByIndex(f.mappings[key].src.Index)
		dv, err := f.mappings[key].format.format(sv)
		if err != nil {
			return reflect.Value{}, err
		}
		dst.FieldByIndex(f.mappings[key].dst.Index).Set(dv)
	}
	return dst, nil
}

type structBuilder struct {
	t        reflect.Type
	builders map[string]builder
	tags     map[string]string
	fields   map[string]reflect.StructField
}

func (b *structBuilder) build(fields []string) (formatter, error) {
	if len(fields) == 0 {
		return &primitiveFormatter{}, nil
	}
	var lf []reflect.StructField
	mappings := map[string]mapping{}
	for _, field := range fields {
		if idx := strings.Index(field, "."); idx != -1 {
			subtag := field[:idx]
			if _, found := mappings[subtag]; found {
				continue
			}
			subbuilder := b.builders[subtag]
			if subbuilder == nil {
				return nil, fmt.Errorf("field '%s' does not exist", subtag)
			}
			var subfields []string
			for _, subfield := range fields {
				if strings.HasPrefix(subfield, field[:idx+1]) {
					subfields = append(subfields, subfield[idx+1:])
				}
			}
			fmter, err := subbuilder.build(subfields)
			if err != nil {
				return nil, err
			}
			sf := reflect.StructField{
				Name:      strings.ToUpper(subtag),
				Tag:       reflect.StructTag(`json:"` + b.tags[subtag] + `"`),
				Type:      fmter.typ(),
				Anonymous: b.fields[subtag].Anonymous,
			}
			lf = append(lf, sf)
			sf.Index = []int{len(lf) - 1}
			mappings[subtag] = mapping{
				src:    b.fields[subtag],
				dst:    sf,
				format: fmter,
			}
		} else {
			fieldbuilder := b.builders[field]
			if fieldbuilder == nil {
				return nil, fmt.Errorf("field '%s' does not exist", field)
			}
			fmter, err := fieldbuilder.build(nil)
			if err != nil {
				return nil, err
			}
			sf := reflect.StructField{
				Name:      strings.ToUpper(field),
				Tag:       reflect.StructTag(`json:"` + b.tags[field] + `"`),
				Type:      fmter.typ(),
				Anonymous: b.fields[field].Anonymous,
			}
			lf = append(lf, sf)
			sf.Index = []int{len(lf) - 1}
			mappings[field] = mapping{
				src:    b.fields[field],
				dst:    sf,
				format: fmter,
			}
		}
	}
	return &structFormatter{t: reflect.StructOf(lf), mappings: mappings}, nil
}

func makeStructBuilder(t reflect.Type) (*structBuilder, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", t.Name())
	}
	sb := structBuilder{
		t:        t,
		builders: map[string]builder{},
		tags:     map[string]string{},
		fields:   map[string]reflect.StructField{},
	}
	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)
		tag := fld.Tag.Get("json")
		if tag == "-" || fld.PkgPath != "" {
			continue
		}
		if tag == "" {
			tag = fld.Name
		}
		field := tag
		if idx := strings.Index(field, ","); idx != -1 {
			field = field[:idx]
		}
		ssb, err := makeBuilder(fld.Type)
		if err != nil {
			return nil, err
		}
		sb.builders[field] = ssb
		sb.tags[field] = tag
		sb.fields[field] = fld
	}
	return &sb, nil
}
