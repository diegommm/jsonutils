// Package jsonutils provides various utilities to handle JSON encoded data.
package jsonutils

import (
	"encoding/json"
	"io"
	"unsafe"
)

// Error is a general purpose and simple error message.
type Error string

func (e Error) Error() string { return string(e) }

// Errors defined by the package.
const (
	ErrEmpty             Error = "empty input"
	ErrUnknownType       Error = "unknown type"
	ErrUnexpectedType    Error = "unexpected JSON type"
	ErrUnexpectedMapping Error = "unexpected mapping"
)

// JSONType identifies one of the stardad JSON Data Types.
type JSONType uint8

// Standard JSON types.
const (
	// InvalidJSON is not a JSON Data Type but it's used for control purposes in
	// this package (like signaling errors or uninitialized values).
	InvalidJSON JSONType = iota

	Object
	Array
	Null
	String
	Number
	Boolean

	// Internal use. Keep at the end of this block.
	maxJSONType
)

// GoMapping specifies how was the JSON Data Type mapped into a Go Type. This
// information is used to understand how can be retrieved the value from the
// Payload once processed.
type GoMapping uint8

const (
	// GoInvalidMapping is not a mapping itself but it's used for control
	// purposes in this package (like signaling errors or uninitialized values).
	GoInvalidMapping GoMapping = iota

	// GoOther means the value can be retrieved as an interface{} with
	// GetObject, GetArray or GetOther. Note that all these three methods are
	// aliases and exist only for mnemonics and convenience of the user. The
	// JSON value was an Object or an Array.
	GoOther
	// GoNil means that the JSON value was Null. Null will be interpreted as
	// nil, but it doesn't make sense to retrieve a nil value. One can safely
	// check if a Payload was Null by calling IsNil.
	GoNil
	// GoString means that the JSON value was a String. The value can be
	// retrieved as a string with GetString.
	GoString
	// GoInt means that the JSON value was a Number and that WithNumber or
	// WithInt were used. The value can be retrieved as an int64 with GetInt.
	GoInt
	// GoFloat means that the JSON value was a Number and that the WithFloat was
	// used. The value can be retrieved as a float64 with GetFloat.
	GoFloat
	// GoUint means that the JSON value was a Number and that the WithUint was
	// used. The value can be retrieved as a float64 with GetUint.
	GoUint
	// GoBool means that the JSON value was a Boolean and that the WithBool was
	// used. The value can be retrieved as a bool with GetBool.
	GoBool
)

func (m GoMapping) panicIfNot(m2 GoMapping) {
	if m != m2 {
		panic(ErrUnexpectedMapping)
	}
}

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

// Embed this type into a struct, which mustn't be copied, so `go vet` gives a
// warning if this struct is copied.
//
// For details, see:
//     https://github.com/golang/go/issues/8005#issuecomment-190753527
//     https://stackoverflow.com/questions/52494458/nocopy-minimal-example
//
type noCopy struct{}    //nolint:unused
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Assert the correctness of our interface that allows mocking.
var _ jsonDecoder = (*json.Decoder)(nil)

var (
	bNull  = []byte{'n', 'u', 'l', 'l'}
	bTrue  = []byte{'t', 'r', 'u', 'e'}
	bFalse = []byte{'f', 'a', 'l', 's', 'e'}
)

// bytesToString converts from []byte to string with no memeroy allocation.
// Caould eventually break if headers for slice or string types are modified.
// See:
// 	https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b)) // #nosec G103
}
