package jsonutils

import (
	"bytes"
	"encoding/json"
	"strconv"
	"sync"
)

var payloadPool sync.Pool

// PayloadFactory serves to document the expected signature for custom factory
// methods. The returned value must be a pointer.
type PayloadFactory = func() interface{}

// WithDefaultArray is the default PayloadFactory for JSON Array.
//
// This simply returns `new([]interface{})`.
func WithDefaultArray() interface{} { return new([]interface{}) }

// WithDefaultObject is the default PayloadFactory for JSON Object.
//
// This simply returns `new(map[string]interface{})`.
func WithDefaultObject() interface{} { return new(map[string]interface{}) }

// Payload allows easy implementation of variable JSON Data Type unmarshalers.
//
// To create a new Payload use the AcquirePayload function and call
// ReleasePayload when it's no longer needed.
//
// Payload by default returns an unmarshaling error for any value. To configure
// it use the With* methods.
//
// Payload has the following restrictions:
//	- It cannot be compared to other Payload object. The comparison should be
//		performed with the _values_, not with the Payload object (use the Get*
//		methods to retrieve the value).
//	- It cannot be copied.
//	- It's not safe to use in concurrent goroutines.
//
// This restrictions aim at improving code quality and performance.
type Payload struct {
	_        noCopy
	with     [maxJSONType]bool
	jsonType JSONType

	// Which of the elements in this struct is holding the value
	mapping GoMapping

	// Boolean Payload
	pBool bool

	// Number Payload
	numType GoMapping // Hint for the unmarshaler on where to put Number values
	pInt    int64
	pUint   uint64
	pFloat  float64

	// String Payload
	pString string

	// Array and Object Payloads
	otherFactory PayloadFactory
	pOther       interface{}
}

// AcquirePayload returns a new Payload from the internal pool.
//
// When no longer used, the Payload should be passed to ReleasePayload to
// recycle the instance, which generally improves performance by reducing the
// load in the garbage collector.
//
// Typically, you would do something like this:
//	p := AcquirePayload()
//	defer ReleasePayload(p)
//
func AcquirePayload() *Payload {
	if p := payloadPool.Get(); p != nil {
		return p.(*Payload)
	}
	return &Payload{}
}

// ReleasePayload releases the resources associated with the Payload. After
// calling this function, the reference associated with this Payload should not
// be used, otherwise unexpected errors could arise.
func ReleasePayload(p *Payload) {
	if p != nil {
		p.Reset()
		payloadPool.Put(p)
	}
}

// Assert at compile-time that we implement the JSON Unmarshaler interface.
var _ json.Unmarshaler = (*Payload)(nil)

// Reset clears all configurations and values. After this operation the object
// gets to an initial state.
func (p *Payload) Reset() {
	p.Clear()
	for i := range p.with {
		p.with[i] = false
	}
	p.otherFactory = nil
	p.numType = GoInvalidMapping
}

// Clear removes all the associated data saved in the Payload but keeping all
// the configurations.
func (p *Payload) Clear() {
	p.jsonType = InvalidJSON
	p.mapping = GoInvalidMapping
	p.pBool = false
	p.pInt = 0
	p.pUint = 0
	p.pFloat = 0
	p.pString = ""
	p.pOther = nil
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (p *Payload) UnmarshalJSON(b []byte) error {
	p.Clear() // Reset state before attempting unmarshal.

	var err error
	if p.jsonType, err = TypeOf(b); err != nil {
		return err
	}

	if p.jsonType == InvalidJSON || !p.with[p.jsonType] {
		return ErrUnexpectedType
	}

	switch p.jsonType {
	case Object, Array:
		p.mapping = GoOther
		p.pOther = p.otherFactory()
		err = json.Unmarshal(b, p.pOther)

	case Null:
		p.mapping = GoNil

	case String:
		p.mapping = GoString
		p.pString, err = strconv.Unquote(bytesToString(b))

	case Number:
		p.mapping = p.numType
		switch p.mapping {
		case GoInt:
			p.pInt, err = strconv.ParseInt(bytesToString(b), 10, 64)
		case GoFloat:
			p.pFloat, err = strconv.ParseFloat(bytesToString(b), 64)
		case GoUint:
			p.pUint, err = strconv.ParseUint(bytesToString(b), 10, 64)
		}

	case Boolean:
		p.mapping = GoBool
		p.pBool = bytes.Compare(bTrue, b) == 0
	}

	if err != nil {
		p.mapping = GoInvalidMapping
	}

	return err
}

// Get retrieves the Payload value as an interface{}.
//
// 	- If the Payload was never loaded (through JSON unmarshaling) nil and
// 		GoInvalidMapping are returned.
//	- If the Payload was loaded but an unmarshaling error occurred (like the
//		Payload not being configured to receive that JSON Data Type or any other
//		unmarshaling error) then it also returns nil and GoInvalidMapping.
// 	- If the Payload was loaded and it was the JSON Null value then nil and
// 		GoNil are returned.
//	- Otherwise, the corresponding value is returned and the GoMapping return
//		value gives a hint on how to interpret that interface{} value.
//
// Get never panics.
func (p *Payload) Get() (interface{}, GoMapping) {
	var ret interface{}
	switch p.mapping {
	case GoOther:
		ret = p.pOther
	case GoString:
		ret = p.pString
	case GoFloat:
		ret = p.pFloat
	case GoInt:
		ret = p.pInt
	case GoUint:
		ret = p.pUint
	case GoBool:
		ret = p.pBool
	}
	return ret, p.mapping
}

// GetObject is an alias for GetOther for conveniency.
func (p *Payload) GetObject() interface{} { return p.GetOther() }

// GetArray is an alias for GetOther for conveniency.
func (p *Payload) GetArray() interface{} { return p.GetOther() }

// GetOther retrieves the Payload value as an interface{}, knowing that the
// JSON Data Type loaded was an Array or an Object.
//
// It panics if the JSON Data Type was not an Array or an Object.
//
func (p *Payload) GetOther() interface{} {
	p.mapping.panicIfNot(GoOther)
	return p.pOther
}

// GetString retrieves the Payload value as a string.
//
// It panics if the JSON Data Type was not a String.
func (p *Payload) GetString() string {
	p.mapping.panicIfNot(GoString)
	return p.pString
}

// GetBool retrieves the Payload value as a bool.
//
// It panics if the JSON Data Type was not an Boolean.
func (p *Payload) GetBool() bool {
	p.mapping.panicIfNot(GoBool)
	return p.pBool
}

// GetInt retrieves the Payload value as an int64.
//
// It panics if the JSON Data Type was not a Number or if the Number mapping was
// meant for a float64 or uint64.
func (p *Payload) GetInt() int64 {
	p.mapping.panicIfNot(GoInt)
	return p.pInt
}

// GetUint retrieves the Payload value as a uint64.
//
// It panics if the JSON Data Type was not a Number or if the Number mapping was
// meant for a float64 or int64.
func (p *Payload) GetUint() uint64 {
	p.mapping.panicIfNot(GoUint)
	return p.pUint
}

// GetFloat retrieves the Payload value as a float64.
//
// It panics if the JSON Data Type was not a Number or if the Number mapping was
// meant for a uint64 or int64.
func (p *Payload) GetFloat() float64 {
	p.mapping.panicIfNot(GoFloat)
	return p.pFloat
}

// IsNil reports whether the Payload value was a JSON Null. It never panics.
//
// Note that if you are using a pointer to Payload the JSON Unmarshaler can
// potentially set it to null in case a JSON Null is unmarshaled so this method
// is not always necessary: just checking that the *Payload is nil would be
// enough and appropriate.
func (p *Payload) IsNil() bool { return p.mapping == GoNil }

// GetMapping returns the Payload value mapping.
func (p *Payload) GetMapping() GoMapping { return p.mapping }

// GetJSONType returns the JSON Data Type associated with the decoded data. It's
// zero value is InvalidJSON, which may also be the value set in case of an
// error when decoding, so make sure to always check unmarshaling errors.
func (p *Payload) GetJSONType() JSONType { return p.jsonType }

// WithNull configures the Payload to accept a JSON Null value (disabled by
// default).
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithNull(enable ...bool) *Payload {
	p.with[Null] = len(enable) == 0 || enable[0]
	return p
}

// WithBoolean configures the Payload to accept a JSON Boolean value (disabled
// by default).
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithBoolean(enable ...bool) *Payload {
	p.with[Boolean] = len(enable) == 0 || enable[0]
	return p
}

// WithString configures the Payload to accept a JSON String value (disabled by
// default).
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithString(enable ...bool) *Payload {
	p.with[String] = len(enable) == 0 || enable[0]
	return p
}

// WithNumber configures the Payload to accept a JSON Number value (disabled by
// default).
//
// When using WithNumber to enable this behavior then it works like calling
// WithInt. Otherwise, JSON Number will be disabled (regardless if it had been
// previously enabled using WithUint or WithFloat).
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithNumber(enable ...bool) *Payload {
	p.with[Number] = len(enable) == 0 || enable[0]
	if p.with[Number] {
		return p.WithInt()
	}
	return p
}

func (p *Payload) withNum(m GoMapping, enable ...bool) *Payload {
	if len(enable) == 0 || enable[0] {
		p.with[Number] = true
		p.numType = m
	} else if p.numType == m && p.with[Number] {
		p.with[Number] = false
	}
	return p
}

// WithFloat configures the Payload to accept a JSON Number value (disabled by
// default) and interpret it as a float64.
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithFloat(enable ...bool) *Payload {
	return p.withNum(GoFloat, enable...)
}

// WithInt configures the Payload to accept a JSON Number value (disabled by
// default) and interpret it as a int64.
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithInt(enable ...bool) *Payload {
	return p.withNum(GoInt, enable...)
}

// WithUint configures the Payload to accept a JSON Number value (disabled by
// default) and interpret it as a uint64.
//
// The default behavior when calling this method is to enable this
// configuration.
func (p *Payload) WithUint(enable ...bool) *Payload {
	return p.withNum(GoUint, enable...)
}

func (p *Payload) withOther(t JSONType, f ...PayloadFactory) *Payload {
	p.with[t] = f[0] != nil
	p.otherFactory = f[0]
	return p
}

// WithArray configures the Payload to accept a JSON Array value (disabled by
// default).
//
//	- When no arguments are passed, WithDefaultArray is assumed.
//	- When a nil argument is passed then this configuration will be disabled.
//	- When a non-nil argument is passed then it will be used to generate a new
//		Go pointer that will be used by the encoding/json Unmarshaler to hold
//		the JSON Array.
func (p *Payload) WithArray(f ...PayloadFactory) *Payload {
	return p.withOther(Array, append(f, WithDefaultArray)...)
}

// WithObject configures the Payload to accept a JSON Object value (disabled by
// default).
//
//	- When no arguments are passed, WithDefaultObject is assumed.
//	- When a nil argument is passed then this configuration will be disabled.
//	- When a non-nil argument is passed then it will be used to generate a new
//		Go pointer that will be used by the encoding/json Unmarshaler to hold
//		the JSON Object.
func (p *Payload) WithObject(f ...PayloadFactory) *Payload {
	return p.withOther(Object, append(f, WithDefaultObject)...)
}
