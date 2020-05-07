package dynjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestFormatter_FormatObject(t *testing.T) {
	var tests = []struct {
		src    interface{}
		format string
		output string
		err    string
	}{
		{
			src:    struct{ Foo int }{Foo: 1},
			format: "bar",
			err:    "no field bar found",
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
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			f := NewFormatter()
			var fields []string
			if tt.format != "" {
				fields = strings.Split(tt.format, ",")
			}
			o, err := f.FormatObject(tt.src, fields, nil)
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

type testEmbed struct {
	Bar int `json:"bar"`
	Baz int `json:"baz"`
}
type testStruct struct {
	Foo int    `json:"foo"`
	Bar string `json:"bar,omitempty"`
}

func testEmbedder(id interface{}) (interface{}, error) {
	switch id.(int) {
	case 1:
		return &testEmbed{Bar: 2, Baz: 3}, nil
	case 2:
		return testEmbed{Bar: 3, Baz: 4}, nil
	default:
		return nil, errors.New("not found")
	}
}

func TestFormatter_Embed(t *testing.T) {
	f := NewFormatter()
	err := f.RegisterEmbed("embed", "foo", testEmbedder, &testEmbed{})
	if err != nil {
		t.Error("Register returned unexpected error", err)
	}
	var tests = []struct {
		foo    int
		embed  string
		format string
		output string
		err    string
	}{
		{
			foo:   1,
			embed: "foo",
			err:   "embed foo was not registered",
		},
		{
			foo:    1,
			embed:  "embed",
			output: `{"foo":1,"embed":{"bar":2,"baz":3}}`,
		},
		{
			foo:    1,
			embed:  "embed",
			format: "foo,embed.bar",
			output: `{"foo":1,"embed":{"bar":2}}`,
		},
		{
			foo:    2,
			embed:  "embed",
			output: `{"foo":2,"embed":{"bar":3,"baz":4}}`,
		},
		{
			foo:   3,
			embed: "embed",
			err:   "not found",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			var fields []string
			if tt.format != "" {
				fields = strings.Split(tt.format, ",")
			}
			o, err := f.FormatObject(testStruct{Foo: tt.foo}, fields, []string{tt.embed})
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

func TestFormatter_FormatList(t *testing.T) {
	f := NewFormatter()
	err := f.RegisterEmbed("embed", "foo", testEmbedder, &testEmbed{})
	if err != nil {
		t.Error("Register returned unexpected error", err)
	}
	var tests = []struct {
		src    []testStruct
		embed  string
		format string
		output string
		err    string
	}{
		{
			src:    []testStruct{{Foo: 1, Bar: "bar"}, {Foo: 2, Bar: "baz"}},
			output: `[{"foo":1,"bar":"bar"},{"foo":2,"bar":"baz"}]`,
		},
		{
			src:    []testStruct{{Foo: 1, Bar: "bar"}, {Foo: 2, Bar: "baz"}},
			format: "foo",
			output: `[{"foo":1},{"foo":2}]`,
		},
		{
			src:    []testStruct{{Foo: 1}, {Foo: 2}},
			embed:  "embed",
			output: `[{"foo":1,"embed":{"bar":2,"baz":3}},{"foo":2,"embed":{"bar":3,"baz":4}}]`,
		},
		{
			src:   []testStruct{{Foo: 1}, {Foo: 3}},
			embed: "embed",
			err:   "not found",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			var fields, embeds []string
			if tt.format != "" {
				fields = strings.Split(tt.format, ",")
			}
			if tt.embed != "" {
				embeds = strings.Split(tt.embed, ",")
			}
			o, err := f.FormatList(tt.src, fields, embeds)
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

func BenchmarkFormatter_FormatObject_Fields(b *testing.B) {
	f := NewFormatter()
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		o, _ := f.FormatObject(testStruct{Foo: i, Bar: "bar"}, []string{"foo", "bar"}, nil)
		w.Encode(o)
	}
}

func BenchmarkFormatter_FormatObject_NoFields(b *testing.B) {
	f := NewFormatter()
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		o, _ := f.FormatObject(testStruct{Foo: i, Bar: "bar"}, nil, nil)
		w.Encode(o)
	}
}

func BenchmarkRawJSON(b *testing.B) {
	w := json.NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		w.Encode(testStruct{Foo: i, Bar: "bar"})
	}
}
