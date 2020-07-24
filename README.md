# dynjson [![](https://godoc.org/github.com/cocoonspace/dynjson?status.svg)](https://godoc.org/github.com/cocoonspace/dynjson) [![Build Status](https://travis-ci.org/cocoonspace/dynjson.svg?branch=master)](https://travis-ci.org/cocoonspace/dynjson) [![Coverage Status](https://coveralls.io/repos/github/cocoonspace/dynjson/badge.svg?branch=master)](https://coveralls.io/github/cocoonspace/dynjson?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/cocoonspace/dynjson)](https://goreportcard.com/report/github.com/cocoonspace/dynjson)

Client-customizable JSON formats for dynamic APIs.


## Introduction

dynjson allow APIs to return only fields selected by the API client :

```
GET https://api.example.com/v1/foos
[{"id":1,foo":1,"bar":2,"baz":3}]

GET https://api.example.com/v1/foos?select=foo
[{"foo":1}]

GET https://api.example.com/v1/foos/1?select=foo
{"foo":1}
```

## Installation

go get github.com/cocoonspace/dynjson

## Usage

```go
type APIResult struct {
    Foo int     `json:"foo"`
    Bar string  `json:"bar"`
}

f := dynjson.NewFormatter()

res := &APIResult{Foo:1, Bar:"bar"}
o, err := f.Format(res, dynjson.FieldsFromRequest(r))
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"foo": 1}
```

With struct fields :


```go
type APIResult struct {
    Foo int          `json:"foo"`
    Bar APIIncluded  `json:"bar"`
}

type APIIncluded struct {
    BarFoo int    `json:"barfoo"`
    BarBar string `json:"barbar"`
}

f := dynjson.NewFormatter()

res := &APIResult{Foo: 1, Bar: APIIncluded{BarFoo:1, BarBar: "bar"}}
o, err := f.Format(res, []string{"foo", "bar.barfoo"})
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"foo": 1, "bar":{"barfoo": 1}}
```

With slices:

```go
type APIResult struct {
    Foo int     `json:"foo"`
    Bar string  `json:"bar"`
}

f := dynjson.NewFormatter()

res := []APIResult{{Foo: 1, Bar: "bar"}}
o, err := f.Format(res, []string{"foo"})
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // [{"foo": 1}]
```


```go
type APIResult struct {
    Foo int        `json:"foo"`
    Bar []APIItem  `json:"bar"`
}

type APIItem struct {
    BarFoo int    `json:"barfoo"`
    BarBar string `json:"barbar"`
}

f := dynjson.NewFormatter()

res := &APIResult{Foo: 1, Bar: []APIItem{{BarFoo: 1, BarBar: "bar"}}}
o, err := f.Format(res, []string{"foo", "bar.barfoo"})
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"foo": 1, "bar":[{"barfoo": 1}]}
```

# Anonymous fields

Anonymous fields without a json tag are not supported.

# Performance impact

```
BenchmarkFormat_Fields
BenchmarkFormat_Fields-8     	 2466639	       480 ns/op	     184 B/op	       7 allocs/op
BenchmarkFormat_NoFields
BenchmarkFormat_NoFields-8   	 5255031	       232 ns/op	      32 B/op	       1 allocs/op
BenchmarkRawJSON
BenchmarkRawJSON-8           	 5351313	       223 ns/op	      32 B/op	       1 allocs/op
```

# License

MIT - See LICENSE