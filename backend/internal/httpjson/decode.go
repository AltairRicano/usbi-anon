package httpjson

import (
	"encoding/json"
	"errors"
	"io"
)

var ErrMultipleJSONValues = errors.New("request body must contain a single JSON value")

func DecodeStrict(r io.Reader, dst any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ErrMultipleJSONValues
		}
		return err
	}

	return nil
}
