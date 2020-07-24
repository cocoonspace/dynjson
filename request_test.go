package dynjson_test

import (
	"net/http"
	"testing"

	"pkgs/dynjson"
)

func TestFieldsFromRequest(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint?select=foo&select=bar", nil)
	if err != nil {
		t.Error("Should not have returned", err)
	}
	fields := dynjson.FieldsFromRequest(r)
	if len(fields) != 2 {
		t.Error("2 fields were expected")
	}
	if fields[0] != "foo" || fields[1] != "bar" {
		t.Errorf("Expected [foo bar] but got %v", fields)
	}

}
