package dynjson

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type mapping struct {
	src   reflect.StructField
	dst   reflect.StructField
	embed string
}

type format struct {
	t      reflect.Type
	fields map[string]mapping
}

type embed struct {
	idfield  string
	embedder Embedder
	t        reflect.Type
}

// Formatter is a dynamic API format formatter.
type Formatter struct {
	mu         sync.Mutex
	fieldnames map[reflect.Type][]string
	fields     map[reflect.Type]map[string]reflect.StructField
	formats    map[reflect.Type]map[string]format
	embeds     map[string]embed
}

// NewFormatter creates a new formatter.
func NewFormatter() *Formatter {
	return &Formatter{
		fieldnames: map[reflect.Type][]string{},
		fields:     map[reflect.Type]map[string]reflect.StructField{},
		formats:    map[reflect.Type]map[string]format{},
		embeds:     map[string]embed{},
	}
}

// Embedder is the func that returns an embedded sub-resource using its id.
type Embedder func(id interface{}) (interface{}, error)

// RegisterEmbed register an embed.
func (f *Formatter) RegisterEmbed(name string, idfield string, e Embedder, i interface{}) error {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f.embeds[name] = embed{
		idfield:  idfield,
		embedder: e,
		t:        t,
	}
	return f.addStruct(t, name+".")
}

// FormatObject formats an object, returning only the selected fields (or all if none specified), including the selected embeds.
func (f *Formatter) FormatObject(o interface{}, fields []string, embeds []string) (interface{}, error) {
	src := reflect.Indirect(reflect.ValueOf(o))
	ff, err := f.getFormat(src.Type(), fields, embeds)
	if err != nil {
		return nil, err
	}
	dst, err := f.doFormatObject(src, ff, fields, embeds)
	if err != nil {
		return nil, err
	}
	return dst.Interface(), nil
}

// FormatList formats a slice.
func (f *Formatter) FormatList(o interface{}, fields []string, embeds []string) (interface{}, error) {
	src := reflect.Indirect(reflect.ValueOf(o))
	if src.Kind() != reflect.Slice {
		return nil, errors.New("input is not a slice")
	}
	ff, err := f.getFormat(src.Type().Elem(), fields, embeds)
	if err != nil {
		return nil, err
	}
	dst := reflect.MakeSlice(reflect.SliceOf(ff.t), src.Len(), src.Len())
	for i := 0; i < src.Len(); i++ {
		out, err := f.doFormatObject(src.Index(i), ff, fields, embeds)
		if err != nil {
			return nil, err
		}
		dst.Index(i).Set(out)
	}
	return dst.Interface(), nil
}

func (f *Formatter) doFormatObject(src reflect.Value, ff format, fields []string, embeds []string) (reflect.Value, error) {
	pdst := reflect.New(ff.t)
	dst := pdst.Elem()
	embedded := map[string]reflect.Value{}
	if len(embeds) > 0 {
		var ids []string
		for _, e := range embeds {
			ids = append(ids, f.embeds[e].idfield)
		}
		def, err := f.getFormat(src.Type(), ids, nil)
		if err != nil {
			return reflect.Value{}, err
		}
		for i, e := range embeds {
			iv := src.FieldByIndex(def.fields[ids[i]].src.Index)
			o, err := f.embeds[e].embedder(iv.Interface())
			if err != nil {
				return reflect.Value{}, err
			}
			embedded[e] = reflect.Indirect(reflect.ValueOf(o))
		}
	}
	for key := range ff.fields {
		var sf reflect.Value
		if ff.fields[key].embed != "" {
			sf = embedded[ff.fields[key].embed].FieldByIndex(ff.fields[key].src.Index)
		} else {
			sf = src.FieldByIndex(ff.fields[key].src.Index)
		}
		dst.FieldByIndex(ff.fields[key].dst.Index).Set(sf)
	}
	return dst, nil
}

func (f *Formatter) getFormat(t reflect.Type, fields []string, embeds []string) (format, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fields[t] == nil {
		if err := f.addStruct(t, ""); err != nil {
			return format{}, err
		}
		f.formats[t] = map[string]format{}
	}
	if len(fields) == 0 {
		fields = f.fieldnames[t]
		for _, embed := range embeds {
			if e, found := f.embeds[embed]; found {
				fields = append(fields, f.fieldnames[e.t]...)
			} else {
				return format{}, fmt.Errorf("embed %s was not registered", embed)
			}
		}
	}
	sig := strings.Join(fields, "|")
	if fmt, found := f.formats[t][sig]; found {
		return fmt, nil
	}
	fmt, err := f.buildFormat(t, fields, embeds)
	if err != nil {
		return format{}, err
	}
	f.formats[t][sig] = fmt
	return fmt, nil
}

func (f *Formatter) addStruct(t reflect.Type, prefix string) error {
	names, fields, err := f.parseStruct(t)
	if err != nil {
		return err
	}
	f.fields[t] = map[string]reflect.StructField{}
	for _, name := range names {
		f.fieldnames[t] = append(f.fieldnames[t], prefix+name)
		f.fields[t][prefix+name] = fields[name]
	}
	return nil
}

func (f *Formatter) parseStruct(t reflect.Type) ([]string, map[string]reflect.StructField, error) {
	if t.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("%s is not a struct", t.Name())
	}
	var fields []string
	sf := map[string]reflect.StructField{}
	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)
		tag := fld.Tag.Get("json")
		if tag == "-" {
			continue
		}
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		if tag == "" {
			tag = fld.Name
		}
		if fld.Type.Kind() == reflect.Struct || (fld.Type.Kind() == reflect.Ptr && fld.Type.Elem().Kind() == reflect.Struct) {
			st := fld.Type
			if st.Kind() == reflect.Ptr {
				st = st.Elem()
			}
			subfields, subsf, err := f.parseStruct(st)
			if err != nil {
				return nil, nil, err
			}
			for _, subfield := range subfields {
				path := tag + "." + subfield
				sf[path] = reflect.StructField{
					Name:      subsf[subfield].Name,
					PkgPath:   subsf[subfield].PkgPath,
					Type:      subsf[subfield].Type,
					Tag:       subsf[subfield].Tag,
					Index:     append([]int{i}, subsf[subfield].Index...),
					Anonymous: subsf[subfield].Anonymous,
				}
				fields = append(fields, path)
			}
		} else {
			sf[tag] = fld
			fields = append(fields, tag)
		}
	}
	return fields, sf, nil
}

func (f *Formatter) buildType(t reflect.Type, fields []string, embeds map[string]bool, prefix string) (reflect.Type, error) {
	done := map[string]bool{}
	var lf []reflect.StructField
	for _, fld := range fields {
		if !strings.HasPrefix(fld, prefix) {
			continue
		}
		stripped := strings.TrimPrefix(fld, prefix)
		if idx := strings.Index(stripped, "."); idx != -1 {
			subtag := fld[:idx]
			if done[subtag] {
				continue
			}
			var (
				subt reflect.Type
				err  error
			)
			if embeds[subtag] {
				subt, err = f.buildType(f.embeds[subtag].t, fields, nil, prefix+subtag+".")
			} else {
				subt, err = f.buildType(t, fields, nil, prefix+subtag+".")
			}
			if err != nil {
				return nil, err
			}
			done[subtag] = true
			lf = append(lf, reflect.StructField{
				Name: strings.ToUpper(subtag),
				Tag:  reflect.StructTag(`json:"` + subtag + `"`),
				Type: subt,
			})
		} else {
			if sf, found := f.fields[t][fld]; found {
				lf = append(lf, sf)
			} else {
				return nil, fmt.Errorf("no field %s found", fld)
			}
		}
	}

	return reflect.StructOf(lf), nil
}

func (f *Formatter) buildFormat(t reflect.Type, fields []string, embeds []string) (format, error) {
	embedsmap := map[string]bool{}
	for _, e := range embeds {
		embedsmap[e] = true
	}
	ot, err := f.buildType(t, fields, embedsmap, "")
	if err != nil {
		return format{}, err
	}
	ff := format{t: ot, fields: map[string]mapping{}}
	_, dsf, err := f.parseStruct(ot)
	if err != nil {
		return format{}, err
	}
	for k, df := range dsf {
		if idx := strings.Index(k, "."); idx != -1 && embedsmap[k[:idx]] {
			name := k[:idx]
			if sf, found := f.fields[f.embeds[name].t][k]; found {
				ff.fields[k] = mapping{
					src: sf, dst: df,
					embed: name,
				}
			} else {
				return format{}, fmt.Errorf("missing field %s in input format", k)
			}
		} else if sf, found := f.fields[t][k]; found {
			ff.fields[k] = mapping{
				src: sf, dst: df,
			}
		} else {
			return format{}, fmt.Errorf("missing field %s in input format", k)
		}
	}
	return ff, nil
}
