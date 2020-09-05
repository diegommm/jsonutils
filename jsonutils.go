// Package jsonutils provides various utilities to handle JSON encoded data.
package jsonutils

import (
	"encoding/json"
	"io"
)

// Error is a general purpose and simple error message.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrEmpty          Error = "empty input"
	ErrUnknownType    Error = "unknown type"
	ErrUnexpectedType Error = "unexpected type"
)

// JSONType identifies one of the stardad JSON Data Types.
type JSONType uint8

const (
	// Invalid is not a JSON Data Type but it's used for control purposes in
	// this package (like signaling errors or uninitialized values).
	Invalid JSONType = iota

	Array
	Boolean
	Null
	Number
	Object
	String

	// Internal use. Keep at the end of this block.
	maxJSONType
)

// Embed this type into a struct, which mustn't be copied, so `go vet` gives a
// warning if this struct is copied.
//
// For details, see:
//	https://github.com/golang/go/issues/8005#issuecomment-190753527
//	https://stackoverflow.com/questions/52494458/nocopy-minimal-example
//
type noCopy struct{} //nolint:unused

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Allow mocking of JSON decoder.
type jsonDecoder interface {
	Buffered() io.Reader
	Decode(interface{}) error
	DisallowUnknownFields()
	InputOffset() int64
	More() bool
	Token() (json.Token, error)
	UseNumber()
}

// Assert the correctness of our interface that allows mocking.
var _ jsonDecoder = (*json.Decoder)(nil)
