package dynjson

import (
	"encoding/json"
	"net/http"
)

func ExampleFormatter_Format() {
	var w http.ResponseWriter
	var r *http.Request

	type APIResult struct {
		Foo int    `json:"foo"`
		Bar string `json:"bar"`
	}

	f := NewFormatter()

	res := &APIResult{Foo: 1, Bar: "bar"}
	o, err := f.Format(res, FieldsFromRequest(r))
	if err != nil {
		// handle error
	}
	err = json.NewEncoder(w).Encode(o) // {"foo": 1}
	if err != nil {
		// handle error
	}
}
