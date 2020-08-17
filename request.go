package dynjson

import (
	"net/http"
	"net/url"
	"strings"
)

// Option defines a FieldsFromRequest option.
type Option int

const (
	// OptionMultipleFields expects multiple select query parameters.
	OptionMultipleFields Option = iota
	// OptionCommaList expects a single select parameter with comma separated values.
	OptionCommaList
)

// FieldsFromRequest returns the list of fields requested from a http.Request.
//
// Without opt or with OptionMultipleFields, the expected format is:
// http://api.example.com/endpoint?select=foo&select=bar
//
// With OptionCommaList, the expected format is:
// http://api.example.com/endpoint?select=foo,bar
func FieldsFromRequest(r *http.Request, opt ...Option) []string {
	vals, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil
	}
	if len(opt) == 1 && opt[0] == OptionCommaList && len(vals["select"]) > 0 {
		return strings.Split(vals["select"][0], ",")
	}
	return vals["select"]
}
