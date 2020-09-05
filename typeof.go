package jsonutils

import (
	"bytes"
	"encoding/json"
)

type typeOfSignature = func([]byte) (JSONType, error)

var typeOfImplem = typeOfFromBytes(func(b []byte) jsonDecoder {
	return json.NewDecoder(bytes.NewBuffer(b))
})

// TypeOf determines the JSON Data Type of the specified []byte. This comes in
// handy when needing to decode variable type responses.
//
func TypeOf(jsonBytes []byte) (JSONType, error) {
	return typeOfImplem(jsonBytes)
}

func typeOfFromBytes(newDecoder func([]byte) jsonDecoder) typeOfSignature {
	return func(jsonBytes []byte) (JSONType, error) {
		if len(jsonBytes) == 0 {
			return Invalid, ErrEmpty
		}

		dec := newDecoder(jsonBytes)

		// See the tip of the next token to avoid decoding expensive values.
		switch jsonBytes[0] {
		case '{':
			return Object, nil
		case '[':
			return Array, nil
		case '"':
			return String, nil
		}

		// Not an expensive value, decode it.
		tk, err := dec.Token()
		if err != nil {
			return Invalid, err
		}

		switch tk.(type) {
		case nil:
			return Null, nil
		case bool:
			return Boolean, nil
		case float64, json.Number:
			return Number, nil
		}

		return Invalid, ErrUnknownType
	}
}
