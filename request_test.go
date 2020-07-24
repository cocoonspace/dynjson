package dynjson

import (
	"net/http"
	"testing"
)

func TestFieldsFromRequest(t *testing.T) {
	{
		r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint", nil)
		if err != nil {
			t.Error("Should not have returned", err)
		}
		fields := FieldsFromRequest(r)
		if len(fields) != 0 {
			t.Error("0 fields were expected")
		}
	}
	{
		r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint?select=foo&select=bar", nil)
		if err != nil {
			t.Error("Should not have returned", err)
		}
		fields := FieldsFromRequest(r)
		if len(fields) != 2 {
			t.Error("2 fields were expected")
		}
		if fields[0] != "foo" || fields[1] != "bar" {
			t.Errorf("Expected [foo bar] but got %v", fields)
		}
	}
	{
		r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint?select=foo&select=bar", nil)
		if err != nil {
			t.Error("Should not have returned", err)
		}
		fields := FieldsFromRequest(r, OptionMultipleFields)
		if len(fields) != 2 {
			t.Error("2 fields were expected")
		}
		if fields[0] != "foo" || fields[1] != "bar" {
			t.Errorf("Expected [foo bar] but got %v", fields)
		}
	}
	{
		r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint?select=foo,bar", nil)
		if err != nil {
			t.Error("Should not have returned", err)
		}
		fields := FieldsFromRequest(r, OptionCommaList)
		if len(fields) != 2 {
			t.Error("2 fields were expected")
		}
		if fields[0] != "foo" || fields[1] != "bar" {
			t.Errorf("Expected [foo bar] but got %v", fields)
		}
	}
	{
		r, err := http.NewRequest(http.MethodGet, "http://api.example.com/endpoint?select=ad%f", nil)
		if err != nil {
			t.Error("Should not have returned", err)
		}
		fields := FieldsFromRequest(r)
		if len(fields) != 0 {
			t.Error("0 fields were expected")
		}
	}
}
