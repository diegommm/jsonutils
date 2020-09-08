package jsonutils

import (
	"encoding/json"
	"testing"
)

type Response struct {
	Status int              `json:"status"`
	Data   json.Unmarshaler `json:"data"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Tags []Tag

// Good practice: assert that we implement the desired interface.
var _ json.Unmarshaler = (*Tags)(nil)

func (t *Tags) UnmarshalJSON(b []byte) error {
	var res interface{}
	*t = Tags{}

	switch jType, err := TypeOf(b); {
	case err != nil:
		return err
	case jType == Object:
		*t = append(*t, Tag{})
		res = &(*t)[0]
	case jType == Array:
		res = (*[]Tag)(t)
	default:
		return ErrUnknownType
	}

	if err := json.Unmarshal(b, res); err != nil {
		return err
	}

	return nil
}

// This generates a new response for a hypothetical endpoint that might returns
// a slice of Tag objects if more than one matches the criteria or a single Tag
// object if only one does. This is a common practice in some APIs.
func NewTagsResponse() *Response {
	return &Response{Data: &Tags{}}
}

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}

type People []Person

// Good practice: assert that we implement the desired interface.
var _ json.Unmarshaler = (*People)(nil)

func (p *People) UnmarshalJSON(b []byte) error {
	var res interface{}
	*p = People{}

	switch jType, err := TypeOf(b); {
	case err != nil:
		return err
	case jType == Object:
		*p = append(*p, Person{})
		res = &(*p)[0]
	case jType == Array:
		res = (*[]Person)(p)
	case jType == String:
		*p = append(*p, Person{})
		res = &(*p)[0].Name
	case jType == Number:
		*p = append(*p, Person{})
		res = &(*p)[0].Age
	default:
		return ErrUnknownType
	}

	if err := json.Unmarshal(b, res); err != nil {
		return err
	}

	return nil
}

// This generates a new response for a hypothetical endpoint that might returns
// a slice of Person objects if more than one matches the criteria or a single
// Person object if only one does. This is a common practice in some APIs.
func NewPeopleResponse() *Response {
	return &Response{Data: &People{}}
}

var TypeOfTestsResponse = []struct {
	Name          string
	Payload       []byte
	Response      interface{}
	Error         string
	MarshaledBack string
}{

	{
		Name: "Tags Null (cannot be controlled through a custom " +
			"unmarshaler)",
		Payload:       []byte(`{"status":200,"data":null}`),
		Response:      NewTagsResponse(),
		Error:         "",
		MarshaledBack: `{"status":200,"data":null}`,
	}, //*/

	{
		Name:          "Tags Object",
		Payload:       []byte(`{"status":200,"data":{"key":"K","value":"V"}}`),
		Response:      NewTagsResponse(),
		Error:         "",
		MarshaledBack: `{"status":200,"data":[{"key":"K","value":"V"}]}`,
	}, //*/

	{
		Name: "Tags Array + empty object in array",
		Payload: []byte(`{"status":200,"data":[{"key":"K","value":"V"},` +
			`{"key":"K1","value":"V2"}, {} ]}`),
		Response: NewTagsResponse(),
		Error:    "",
		MarshaledBack: `{"status":200,"data":[{"key":"K","value":"V"},` +
			`{"key":"K1","value":"V2"},{"key":"","value":""}]}`,
	}, //*/

	{
		Name:     "People String",
		Payload:  []byte(`{"status":200,"data":"Some human"}`),
		Response: NewPeopleResponse(),
		Error:    "",
		MarshaledBack: `{"status":200,"data":[{"name":"Some human",` +
			`"email":""}]}`,
	}, //*/

	{
		Name:          "People - Unknown type",
		Payload:       []byte(`{"status":200,"data":true}`),
		Response:      NewPeopleResponse(),
		Error:         ErrUnknownType.Error(),
		MarshaledBack: `{"status":200,"data":[]}`,
	}, //*/

	{
		Name:     "People - Number",
		Payload:  []byte(`{"status":200,"data":35}`),
		Response: NewPeopleResponse(),
		Error:    "",
		MarshaledBack: `{"status":200,"data":[{"name":"","email":"",` +
			`"age":35}]}`,
	}, //*/

	/* Template for Tags Response
	{
		Name:          "",
		Payload:       []byte(`{"status":200,"data": }`),
		Response:   NewTagsResponse(),
		Error:         "",
		MarshaledBack: ``,
	}, //*/

	/* Template for People Response
	{
		Name:          "",
		Payload:       []byte(`{"status":200,"data": }`),
		Response:   NewPeopleResponse(),
		Error:         "",
		MarshaledBack: ``,
	}, //*/

}

func TestTypeOf_Response(t *testing.T) {
	t.Parallel()
	for i := range TypeOfTestsResponse {
		t.Run(TypeOfTestsResponse[i].Name, func(i int) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				var strErr string
				r := TypeOfTestsResponse[i].Response
				err := json.Unmarshal(TypeOfTestsResponse[i].Payload, r)
				if err != nil {
					strErr = err.Error()
				}
				if strErr != TypeOfTestsResponse[i].Error {
					t.Fatalf("Unexpected marshal error\nWant Error: %s"+
						"\n Got Error: %s", TypeOfTestsResponse[i].Error,
						strErr)
				}
				b, err := json.Marshal(r)
				if err != nil {
					t.Fatalf("Unexpected error while marshaling: %v", err)
				}
				if string(b) != TypeOfTestsResponse[i].MarshaledBack {
					t.Fatalf("Failed marshaling back\nWant MarshalBack: %s\n"+
						" Got MarshalBack: %s",
						TypeOfTestsResponse[i].MarshaledBack, b)
				}
			}
		}(i))
	}
}

var TypeOfTestsRow = []struct {
	Name    string
	Payload []byte
	JSONType
	Error string
}{

	{
		Name:     "Empty payload",
		Payload:  nil,
		JSONType: InvalidJSON,
		Error:    ErrEmpty.Error(),
	}, //*/

	{
		Name:     "Object (not even validated)",
		Payload:  json.RawMessage(`{`),
		JSONType: Object,
		Error:    ``,
	}, //*/

	{
		Name:     "Array (Not even validated)",
		Payload:  json.RawMessage(`[`),
		JSONType: Array,
		Error:    ``,
	}, //*/

	{
		Name:     "String (Not even validated)",
		Payload:  json.RawMessage(`"`),
		JSONType: String,
		Error:    ``,
	}, //*/

	{
		Name:     "Invalid Token",
		Payload:  json.RawMessage(`!`),
		JSONType: InvalidJSON,
		Error:    `unknown type`,
	}, //*/

	{
		Name:     "Null",
		Payload:  json.RawMessage(`null`),
		JSONType: Null,
		Error:    ``,
	}, //*/

	{
		Name:     "Boolean",
		Payload:  json.RawMessage(`true`),
		JSONType: Boolean,
		Error:    ``,
	}, //*/

	{
		Name:     "Number (negative int)",
		Payload:  json.RawMessage(`-1`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name:     "Number (positive int)",
		Payload:  json.RawMessage(`1`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name:     "Number (float)",
		Payload:  json.RawMessage(`3.141592`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name: "Max Float64",
		Payload: json.RawMessage(`1.797693134862315708145274237317043567981` +
			`e+308`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name: "Smallest non-zero Float64",
		Payload: json.RawMessage(`4.940656458412465441765687928682213723651` +
			`e-324`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name:     "Min Int64",
		Payload:  json.RawMessage(`-9223372036854775808`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name:     "Max Int64",
		Payload:  json.RawMessage(`9223372036854775807`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	{
		Name:     "Max Uint64",
		Payload:  json.RawMessage(`18446744073709551615`),
		JSONType: Number,
		Error:    ``,
	}, //*/

	/* Template
	{
		Name:     "",
		Payload:  json.RawMessage(``),
		JSONType: InvalidJSON,
		Error:    ``,
	}, //*/

}

func TestTypeOf_Raw(t *testing.T) {
	t.Parallel()
	for i := range TypeOfTestsRow {
		t.Run(TypeOfTestsRow[i].Name, func(i int) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				var strErr string
				jType, err := TypeOf(TypeOfTestsRow[i].Payload)
				if err != nil {
					strErr = err.Error()
				}
				if strErr != TypeOfTestsRow[i].Error {
					t.Fatalf("Unexpected marshal error\nWant Error: %s"+
						"\n Got Error: %s", TypeOfTestsRow[i].Error, strErr)
				}
				if jType != TypeOfTestsRow[i].JSONType {
					t.Fatalf("Unexpected JSON Data Type\nWant Type: %d"+
						"\n Got Type: %d", TypeOfTestsRow[i].JSONType, jType)
				}
			}
		}(i))
	}
}
