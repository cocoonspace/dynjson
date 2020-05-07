# dynjson
User-customizable JSON formats for dynamic APIs.

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

Sub ressources can be embedded in each resource object :
```
GET https://api.example.com/v1/foos?select=foo,bar.baz&embed=bar
[{"foo":1,"bar":{"baz":3}}]

GET https://api.example.com/v1/foos/1?select=foo,bar.baz&embed=bar
{"foo":1,"bar":{"baz":3}}
```

## Installation

go get github.com/cocoon-space/dynjson

## Usage

Standard use :

```go
type APIResult struct {
    Foo int     `json:"foo"`
    Bar string  `json:"bar"`
}

f := dynjson.NewFormatter()

res := &APIResult{Foo:1, Bar:"bar"}
o, err := f.FormatObject(o, []string{"foo"}, nil)
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"foo": 1}
```

Add embeds :

```go
type SubResource struct {
    ID int      `json:"id"`
    Foo string  `json:"foo"`
}
type APIResult struct {
    SubResourceID   int     `json:"sub_ressource_id"`
    Bar             string  `json:"bar"`
}

f := dynjson.NewFormatter()
err := f.RegisterEmbed("sub_resource", func(id interface{}) (interface{}, error) {
    // fetch sub ressource
}, &SubResource{})
if err != nil {
    // handle error
}
res := &APIResult{SubResourceID:1, Bar:"bar"}
o, err := f.FormatObject(o, nil, "sub_resource")
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"sub_resource_id":1,"bar",:"bar","sub_resource":{"id":1,"foo":"foo"}}
```

Filter embeds fields :

```go
type SubResource struct {
    ID int      `json:"id"`
    Foo string  `json:"foo"`
}
type APIResult struct {
    SubResourceID   int     `json:"sub_resource_id"`
    Bar             string  `json:"bar"`
}

f := dynjson.NewFormatter()
err := f.RegisterEmbed("sub_resource", func(id interface{}) (interface{}, error) {
    // fetch sub ressource
}, &SubResource{})
if err != nil {
    // handle error
}
res := &APIResult{SubResourceID:1, Bar:"bar"}
o, err := f.FormatObject(o, []string{"bar", "sub_resource.foo"}, "sub_resource")
if err != nil {
    // handle error
}
err := json.Marshal(w, o) // {"bar",:"bar","sub_resource":{"foo":"foo"}}
```