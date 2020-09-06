package jsonutils

import (
	"bytes"
	"strconv"
)

var (
	bNull  = []byte{'n', 'u', 'l', 'l'}
	bTrue  = []byte{'t', 'r', 'u', 'e'}
	bFalse = []byte{'f', 'a', 'l', 's', 'e'}
)

// TypeOf determines the JSON Data Type of the specified []byte. This comes in
// handy when needing to decode variable type responses.
//
func TypeOf(jsonBytes []byte) (JSONType, error) {
	if len(jsonBytes) == 0 {
		return Invalid, ErrEmpty
	}

	// See the tip of the next token to avoid decoding expensive values.
	switch jsonBytes[0] {
	case '{':
		return Object, nil
	case '[':
		return Array, nil
	case '"':
		return String, nil
	}

	// Other small and efficient, byte-optimized comparisons.
	switch 0 {
	case bytes.Compare(bNull, jsonBytes):
		return Null, nil
	case bytes.Compare(bTrue, jsonBytes), bytes.Compare(bFalse, jsonBytes):
		return Boolean, nil
	}

	// Should be a number then.
	if _, err := strconv.ParseFloat(string(jsonBytes), 64); err == nil {
		return Number, nil
	}

	return Invalid, ErrUnknownType
}
