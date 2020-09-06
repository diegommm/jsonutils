package jsonutils

import "encoding/json"

// PayloadFactory serves to document the expected signature for custom factory
// methods. The returned value must be a non-pointer or nil (only for Null
// handling).
type PayloadFactory = func() interface{}

// Standard factories.
func withNull() interface{}    { return nil }
func withBoolean() interface{} { return new(bool) }
func withFloat() interface{}   { return new(float64) }
func withInt() interface{}     { return new(int64) }
func withUint() interface{}    { return new(uint64) }
func withString() interface{}  { return new(string) }
func withArray() interface{}   { return new([]interface{}) }
func withObject() interface{}  { return new(map[string]interface{}) }

// Payload allows easy implementation of multiple JSON Data Types unmarshalers.
//
// Restrictions:
//	- Copying and passing by value is not allowed. Use pointers.
//	- Comparing is not allowed.
//	- Concurrent use in different goroutines is not allowed.
//
// These restrictions are designed to improve code quality rather than being
// strictly necessary to provide the desired functionality in this specific
// scenario. It's not either that using pointers and not copying dereferenced
// values is better in all cases, but in this case it's better.
//
type Payload struct {
	_        noCopy
	factory  [maxJSONType]PayloadFactory
	jsonType JSONType
	Data     interface{}
}

// Assert at compile-time that we implement the JSON Unmarshaler interface.
var _ json.Unmarshaler = (*Payload)(nil)

// GetJSONType returns the JSON Data Type associated with the decoded data. It's
// zero value is Invalid, which may also be the value set in case of an error
// when decoding, so make sure to always check unmarshaling errors.
func (p *Payload) GetJSONType() JSONType { return p.jsonType }

func (p *Payload) with(t JSONType, defaultF PayloadFactory, f ...PayloadFactory,
) *Payload {
	if len(f) > 0 {
		p.factory[t] = f[0]
	} else {
		p.factory[t] = defaultF
	}
	return p
}

// WithNull specifies that it is acceptable for the payload to be null. A
// specific type of nil type can be returned if desired.
//
// If no arguments are specified, then the default nil interface{} is returned
// when decoding a nil value.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
func (p *Payload) WithNull(f ...PayloadFactory) *Payload {
	return p.with(Null, withNull, f...)
}

// WithBoolean specifies that it is acceptable for the payload to be a boolean.
// A specific payload can be set to override the default *bool returned.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
func (p *Payload) WithBoolean(f ...PayloadFactory) *Payload {
	return p.with(Boolean, withBoolean, f...)
}

// WithNumber specifies that it is acceptable for the payload to be a number.
// A specific payload can be set to override the default *float64 returned.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
// This method is an alias of the WithFloat method.
//
// Note that the following methods are mutually exclusive, calling more than one
// of them will make only the last one take effect (this is a consequence of
// JSON Number being the only numeric type):
//
//	- WithNumber
//	- WithFloat
//	- WithInt
//	- WithUint
//
func (p *Payload) WithNumber(f ...PayloadFactory) *Payload {
	return p.with(Number, withFloat, f...)
}

// WithFloat is an alias of the WithNumber method.
//
// This method doesn't allow to to specify an alternative value. To do so, use
// the WithNumber method.
//
func (p *Payload) WithFloat() *Payload {
	return p.with(Number, withFloat)
}

// WithInt specifies that it is acceptable for the payload to be a number and
// that that number should be an integer which will be mapped to an *int64.
//
// This method doesn't allow to to specify an factory method. To do so, use
// the WithNumber method.
//
// Note that the following methods are mutually exclusive, calling more than one
// of them will make only the last one take effect (this is a consequence of
// JSON Number being the only numeric type):
//
//	- WithNumber
//	- WithFloat
//	- WithInt
//	- WithUint
//
func (p *Payload) WithInt() *Payload {
	return p.with(Number, withInt)
}

// WithUint specifies that it is acceptable for the payload to be a number and
// that that number should be a non-negative integer which will be mapped to an
// *uint64.
//
// This method doesn't allow to to specify an factory method. To do so, use
// the WithNumber method.
//
// Note that the following methods are mutually exclusive, calling more than one
// of them will make only the last one take effect (this is a consequence of
// JSON Number being the only numeric type):
//
//	- WithNumber
//	- WithFloat
//	- WithInt
//	- WithUint
//
func (p *Payload) WithUint() *Payload {
	return p.with(Number, withUint)
}

// WithString specifies that it is acceptable for the payload to be a string. A
// specific payload can be set to override the default *string returned.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
func (p *Payload) WithString(f ...PayloadFactory) *Payload {
	return p.with(String, withString, f...)
}

// WithArray specifies that it is acceptable for the payload to be an array. A
// specific payload can be set to override the default *[]interface{} returned.
// This could be used to return a specific slice type, e.g., *[]int64, *[]bool,
// *[]MyStruct, etc.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
func (p *Payload) WithArray(f ...PayloadFactory) *Payload {
	return p.with(Array, withArray, f...)
}

// WithObject specifies that it is acceptable for the payload to be an object. A
// specific payload can be set to override the default *map[string]interface{}
// returned. This could be used to return a specific object type, e.g.,
// *MyStruct.
//
// If a nil argument is specified, then the default behavior is restored (in
// case of wanting to clear a previously set custom behavior).
//
func (p *Payload) WithObject(f ...PayloadFactory) *Payload {
	return p.with(Object, withObject, f...)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (p *Payload) UnmarshalJSON(b []byte) error {
	var err error
	p.Data, p.jsonType = nil, Invalid

	p.jsonType, err = TypeOf(b)
	if err != nil {
		return err
	}

	if p.factory[p.jsonType] == nil {
		return ErrUnexpectedType
	}

	p.Data = p.factory[p.jsonType]()
	if p.Data == nil {
		if p.jsonType != Null {
			return ErrUnexpectedType
		}
		return nil
	}

	if err = json.Unmarshal(b, p.Data); err != nil {
		return err
	}

	return nil
}
