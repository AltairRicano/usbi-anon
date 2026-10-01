package httpjson

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeStrictRejectsUnknownFields(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}

	err := DecodeStrict(strings.NewReader(`{"name":"USBI","extra":true}`), &dst)
	if err == nil {
		t.Fatal("DecodeStrict() error = nil, want unknown field error")
	}
}

func TestDecodeStrictRejectsMultipleJSONValues(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}

	err := DecodeStrict(strings.NewReader(`{"name":"USBI"} {"name":"extra"}`), &dst)
	if !errors.Is(err, ErrMultipleJSONValues) {
		t.Fatalf("DecodeStrict() error = %v, want %v", err, ErrMultipleJSONValues)
	}
}
