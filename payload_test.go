package jsonutils

import (
	"encoding/json"
	"testing"
)

var TestsPayload_Raw = []struct {
	Name          string
	JSONData      []byte
	Payload       *Payload
	Error         string
	MarshaledBack string
	JSONType
}{

	{
		Name:          "Null payload",
		JSONData:      []byte(`null`),
		Payload:       new(Payload).WithNull(),
		Error:         "",
		MarshaledBack: "null",
		JSONType:      Null,
	}, //*/

	{
		Name:          "Boolean payload",
		JSONData:      []byte(`true`),
		Payload:       new(Payload).WithBoolean(),
		Error:         "",
		MarshaledBack: "true",
		JSONType:      Boolean,
	}, //*/

	{
		Name:          "Number using WithFloat",
		JSONData:      []byte(`3.141592`),
		Payload:       new(Payload).WithFloat(),
		Error:         "",
		MarshaledBack: "3.141592",
		JSONType:      Number,
	}, //*/

	{
		Name:          "Number using WithInt",
		JSONData:      []byte(`-20`),
		Payload:       new(Payload).WithInt(),
		Error:         "",
		MarshaledBack: "-20",
		JSONType:      Number,
	}, //*/

	{
		Name:     "Number using WithUint, failing since float received",
		JSONData: []byte(`34.1`),
		Payload:  new(Payload).WithUint(),
		Error: "json: cannot unmarshal number 34.1 into Go value of" +
			" type uint64",
		MarshaledBack: "0",
		JSONType:      Number,
	}, //*/

	{
		Name: "Number using WithNumber and custom data (float32 instead " +
			"of float64)",
		JSONData: []byte(`72.9`),
		Payload: new(Payload).WithNumber(func() interface{} {
			return new(float32)
		}),
		Error:         "",
		MarshaledBack: "72.9",
		JSONType:      Number,
	}, //*/

	{
		Name:          "String payload",
		JSONData:      []byte(`"Lorem ipsum"`),
		Payload:       new(Payload).WithString(),
		Error:         "",
		MarshaledBack: `"Lorem ipsum"`,
		JSONType:      String,
	}, //*/

	{
		Name:          "Array payload",
		JSONData:      []byte(`[1,2,3,"Lorem ipsum"]`),
		Payload:       new(Payload).WithArray(),
		Error:         "",
		MarshaledBack: `[1,2,3,"Lorem ipsum"]`,
		JSONType:      Array,
	}, //*/

	{
		Name:          "Object payload",
		JSONData:      []byte(`{"name":"John"}`),
		Payload:       new(Payload).WithObject(),
		Error:         "",
		MarshaledBack: `{"name":"John"}`,
		JSONType:      Object,
	}, //*/

	{
		Name:          "Decodeing error propagation",
		JSONData:      []byte(`whatever`),
		Payload:       new(Payload),
		Error:         "invalid character 'w' looking for beginning of value",
		MarshaledBack: `null`,
		JSONType:      Invalid,
	}, //*/

	{
		Name:          "Unexpected type",
		JSONData:      []byte(`true`),
		Payload:       new(Payload),
		Error:         ErrUnexpectedType.Error(),
		MarshaledBack: `null`,
		JSONType:      Boolean,
	}, //*/

	{
		Name:     "Unexpected nil value returned by factory",
		JSONData: []byte(`true`),
		Payload: new(Payload).WithBoolean(func() interface{} {
			return nil
		}),
		Error:         ErrUnexpectedType.Error(),
		MarshaledBack: `null`,
		JSONType:      Boolean,
	}, //*/

	/* Template
	{
		Name:          "",
		JSONData:      []byte(``),
		Payload:       new(Payload),
		Error:         "",
		MarshaledBack: ``,
		JSONType:      Invalid,
	}, //*/

}

func TestPayload_Raw(t *testing.T) {
	t.Parallel()
	for i := range TestsPayload_Raw {
		t.Run(TestsPayload_Raw[i].Name, func(i int) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				var strErr string
				err := TestsPayload_Raw[i].Payload.UnmarshalJSON(
					TestsPayload_Raw[i].JSONData)
				if err != nil {
					strErr = err.Error()
				}
				if strErr != TestsPayload_Raw[i].Error {
					t.Fatalf("Unexpected marshal error\nWant Error: %s"+
						"\n Got Error: %s", TestsPayload_Raw[i].Error, strErr)
				}
				if TestsPayload_Raw[i].JSONType != TestsPayload_Raw[i].Payload.
					GetJSONType() {
					t.Fatalf("Unexpected JSON Data Type\nWant Type: %d"+
						"\n Got Type: %d", TestsPayload_Raw[i].JSONType,
						TestsPayload_Raw[i].Payload.jsonType)
				}
				b, err := json.Marshal(TestsPayload_Raw[i].Payload.Data)
				if err != nil {
					t.Fatalf("Unexpected error in marshal back: %v", err)
				}
				if string(b) != TestsPayload_Raw[i].MarshaledBack {
					t.Fatalf("Failed marshaling back\nWant MarshalBack: %s\n"+
						" Got MarshalBack: %s",
						TestsPayload_Raw[i].MarshaledBack, b)
				}
			}
		}(i))
	}
}
