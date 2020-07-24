package dynjson

import (
	"net/http"
	"net/url"
)

// FieldsFromRequest returns the list of fields requested from a http.Request
// http://api.example.com/endpoint?select=foo&select=bar
func FieldsFromRequest(r *http.Request) []string {
	vals, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil
	}
	return vals["select"]
}
