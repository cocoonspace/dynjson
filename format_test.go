package dynjson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	var tests = []struct {
		src    interface{}
		format string
		output string
		err    string
	}{
		{
			src:    struct{ Foo int }{Foo: 1},
			format: "bar",
			err:    "field 'bar' does not exist",
		},
		{
			src:    struct{ Foo int }{Foo: 1},
			format: "",
			output: `{"Foo":1}`,
		},
		{
			src:    struct{ Foo int }{Foo: 1},
			format: "Foo",
			output: `{"Foo":1}`,
		},
		{
			src: struct {
				Foo int `json:"foo"`
			}{Foo: 1},
			format: "",
			output: `{"foo":1}`,
		},
		{
			src: struct {
				Foo int `json:"foo"`
			}{Foo: 1},
			format: "foo",
			output: `{"foo":1}`,
		},
		{
			src: struct {
				Foo int `json:"foo,omitempty"`
			}{},
			format: "foo",
			output: `{}`,
		},
		{
			src: struct {
				Foo int    `json:"foo"`
				Bar string `json:"-"`
			}{Foo: 1},
			format: "",
			output: `{"foo":1}`,
		},
		{
			src: struct {
				Foo int    `json:"foo"`
				Bar string `json:"bar"`
			}{Foo: 1, Bar: "bar"},
			format: "foo",
			output: `{"foo":1}`,
		},
		{
			src: struct {
				Foo int    `json:"foo"`
				Bar string `json:"bar"`
			}{Foo: 1, Bar: "bar"},
			format: "",
			output: `{"foo":1,"bar":"bar"}`,
		},
		{
			src: struct {
				Foo int    `json:"foo"`
				Bar string `json:"bar"`
			}{Foo: 1, Bar: "bar"},
			format: "foo,bar",
			output: `{"foo":1,"bar":"bar"}`,
		},
		{
			src: struct {
				Foo int    `json:"foo"`
				Bar string `json:"bar"`
			}{Foo: 1, Bar: "bar"},
			format: "bar,foo",
			output: `{"bar":"bar","foo":1}`,
		},
		{
			src: struct {
				Foo struct {
					Bar int    `json:"bar"`
					Baz string `json:"baz"`
				} `json:"foo"`
			}{
				Foo: struct {
					Bar int    `json:"bar"`
					Baz string `json:"baz"`
				}{
					Bar: 1,
					Baz: "baz",
				},
			},
			format: "",
			output: `{"foo":{"bar":1,"baz":"baz"}}`,
		},
		{
			src: struct {
				Foo struct {
					Bar int    `json:"bar"`
					Baz string `json:"baz"`
				} `json:"foo"`
			}{
				Foo: struct {
					Bar int    `json:"bar"`
					Baz string `json:"baz"`
				}{
					Bar: 1,
					Baz: "baz",
				},
			},
			format: "foo.bar",
			output: `{"foo":{"bar":1}}`,
		},
		{
			src: struct {
				Foo struct {
					Bar int `json:"bar"`
				} `json:"foo"`
				Baz string `json:"baz"`
			}{
				Foo: struct {
					Bar int `json:"bar"`
				}{
					Bar: 1,
				},
				Baz: "baz",
			},
			format: "",
			output: `{"foo":{"bar":1},"baz":"baz"}`,
		},
		{
			src: struct {
				Foo struct {
					Bar int `json:"bar"`
				} `json:"foo"`
				Baz string `json:"baz"`
			}{
				Foo: struct {
					Bar int `json:"bar"`
				}{
					Bar: 1,
				},
				Baz: "baz",
			},
			format: "foo.bar,baz",
			output: `{"foo":{"bar":1},"baz":"baz"}`,
		},
		{
			src: struct {
				Foo *struct {
					Bar int `json:"bar"`
				} `json:"foo,omitempty"`
				Baz string `json:"baz"`
			}{
				Foo: &struct {
					Bar int `json:"bar"`
				}{
					Bar: 1,
				},
				Baz: "baz",
			},
			format: "foo.bar,baz",
			output: `{"foo":{"bar":1},"baz":"baz"}`,
		},
		{
			src: struct {
				Foo *struct {
					Bar int `json:"bar"`
				} `json:"foo,omitempty"`
				Baz string `json:"baz"`
			}{
				Foo: nil,
				Baz: "baz",
			},
			format: "foo.bar,baz",
			output: `{"baz":"baz"}`,
		},
		{
			src: struct {
				Foo struct {
					Foo int `json:"foo"`
					Bar int `json:"bar"`
				} `json:"foo"`
				Baz string `json:"baz"`
			}{
				Foo: struct {
					Foo int `json:"foo"`
					Bar int `json:"bar"`
				}{
					Foo: 1,
					Bar: 2,
				},
				Baz: "baz",
			},
			format: "foo.bar,baz,foo.foo",
			output: `{"foo":{"bar":2,"foo":1},"baz":"baz"}`,
		},
		{
			src: struct {
				Foo struct {
					Foo int `json:"foo"`
					Bar int `json:"bar"`
				} `json:"foo"`
				Baz string `json:"baz"`
			}{
				Foo: struct {
					Foo int `json:"foo"`
					Bar int `json:"bar"`
				}{
					Foo: 1,
					Bar: 2,
				},
				Baz: "baz",
			},
			format: "foo.bar,baz,foo.foo",
			output: `{"foo":{"bar":2,"foo":1},"baz":"baz"}`,
		},
		{
			src: []struct {
				Foo int `json:"foo"`
				Bar int `json:"bar"`
			}{
				{
					Foo: 1,
					Bar: 2,
				},
			},
			format: "foo",
			output: `[{"foo":1}]`,
		},
		{
			src: struct {
				Foo []struct {
					Bar int `json:"bar"`
					Baz int `json:"baz"`
				} `json:"foo"`
			}{
				Foo: []struct {
					Bar int `json:"bar"`
					Baz int `json:"baz"`
				}{{
					Bar: 1,
					Baz: 2,
				}},
			},
			format: "foo.bar",
			output: `{"foo":[{"bar":1}]}`,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			f := NewFormatter()
			var fields []string
			if tt.format != "" {
				fields = strings.Split(tt.format, ",")
			}
			o, err := f.Format(tt.src, fields)
			if tt.err != "" {
				if err == nil {
					t.Fail()
				}
				if tt.err != err.Error() {
					t.Errorf("Returned error '%v', expected '%s'", err, tt.err)
				}
			} else {
				if err != nil {
					t.Error("Should not have returned", err)
				}
				buf, err := json.Marshal(o)
				if err != nil {
					t.Error("Should not have returned", err)
				}
				if tt.output != string(buf) {
					t.Errorf("Returned '%s', expected '%s'", string(buf), tt.output)
				}
			}
		})
	}
}

func BenchmarkFormat_Fields(b *testing.B) {
	f := NewFormatter()
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		o, _ := f.Format(struct {
			Foo int
			Bar string
		}{Foo: i, Bar: "bar"}, []string{"foo", "bar"})
		_ = w.Encode(o)
	}
}

func BenchmarkFormat_NoFields(b *testing.B) {
	f := NewFormatter()
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		o, _ := f.Format(struct {
			Foo int
			Bar string
		}{Foo: i, Bar: "bar"}, nil)
		_ = w.Encode(o)
	}
}

func BenchmarkRawJSON(b *testing.B) {
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		_ = w.Encode(struct {
			Foo int
			Bar string
		}{Foo: i, Bar: "bar"})
	}
}
