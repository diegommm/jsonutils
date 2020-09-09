package jsonutils

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var emptyPayload = fmt.Sprintf("%#v", &Payload{})

type PayloadTest struct {
	Name          string
	JSONData      []byte
	Payload       *Payload
	Error         string
	MarshaledBack string
	JSONType
	GoMapping
}

var TestsPayloadRaw = []PayloadTest{

	{
		Name:          "Null payload",
		JSONData:      []byte(`null`),
		Payload:       AcquirePayload().WithNull(),
		Error:         "",
		MarshaledBack: "null",
		JSONType:      Null,
		GoMapping:     GoNil,
	}, //*/

	{
		Name:          "Boolean payload",
		JSONData:      []byte(`true`),
		Payload:       AcquirePayload().WithBoolean(),
		Error:         "",
		MarshaledBack: "true",
		JSONType:      Boolean,
		GoMapping:     GoBool,
	}, //*/

	{
		Name:          "Number using WithFloat",
		JSONData:      []byte(`3.141592`),
		Payload:       AcquirePayload().WithFloat(),
		Error:         "",
		MarshaledBack: "3.141592",
		JSONType:      Number,
		GoMapping:     GoFloat,
	}, //*/

	{
		Name:          "Number using WithInt",
		JSONData:      []byte(`-20`),
		Payload:       AcquirePayload().WithInt(),
		Error:         "",
		MarshaledBack: "-20",
		JSONType:      Number,
		GoMapping:     GoInt,
	}, //*/

	{
		Name:          "Number using WithUint, failing since float received",
		JSONData:      []byte(`34.1`),
		Payload:       AcquirePayload().WithUint(),
		Error:         "34.1",
		MarshaledBack: "null",
		JSONType:      Number,
		GoMapping:     GoInvalidMapping,
	}, //*/

	{
		Name:          "String payload",
		JSONData:      []byte(`"Lorem ipsum"`),
		Payload:       AcquirePayload().WithString(),
		Error:         "",
		MarshaledBack: `"Lorem ipsum"`,
		JSONType:      String,
		GoMapping:     GoString,
	}, //*/

	{
		Name:          "Array payload",
		JSONData:      []byte(`[1,2,3,"Lorem ipsum"]`),
		Payload:       AcquirePayload().WithArray(),
		Error:         "",
		MarshaledBack: `[1,2,3,"Lorem ipsum"]`,
		JSONType:      Array,
		GoMapping:     GoOther,
	}, //*/

	{
		Name:          "Object payload",
		JSONData:      []byte(`{"name":"John"}`),
		Payload:       AcquirePayload().WithObject(),
		Error:         "",
		MarshaledBack: `{"name":"John"}`,
		JSONType:      Object,
		GoMapping:     GoOther,
	}, //*/

	{
		Name:          "Decoding error propagation",
		JSONData:      []byte(`whatever`),
		Payload:       AcquirePayload(),
		Error:         "unknown type",
		MarshaledBack: `null`,
		JSONType:      InvalidJSON,
		GoMapping:     GoInvalidMapping,
	}, //*/

	{
		Name:          "Unexpected type",
		JSONData:      []byte(`true`),
		Payload:       AcquirePayload(),
		Error:         ErrUnexpectedType.Error(),
		MarshaledBack: `null`,
		JSONType:      Boolean,
		GoMapping:     GoInvalidMapping,
	}, //*/

	{
		Name:          "Uint Payload",
		JSONData:      []byte(`6543`),
		Payload:       AcquirePayload().WithUint(),
		Error:         "",
		MarshaledBack: `6543`,
		JSONType:      Number,
		GoMapping:     GoUint,
	}, //*/

	{
		Name:          "WithNumber",
		JSONData:      []byte(`-34567`),
		Payload:       AcquirePayload().WithNumber(),
		Error:         "",
		MarshaledBack: `-34567`,
		JSONType:      Number,
		GoMapping:     GoInt,
	}, //*/

	{
		Name:          "WithUint then remove using WithNumber",
		JSONData:      []byte(`-2`),
		Payload:       AcquirePayload().WithUint().WithNumber(false),
		Error:         ErrUnexpectedType.Error(),
		MarshaledBack: `null`,
		JSONType:      Number,
		GoMapping:     GoInvalidMapping,
	}, //*/

	{
		Name:          "WithFloat then deactivate it",
		JSONData:      []byte(`3.14`),
		Payload:       AcquirePayload().WithFloat(true).WithFloat(false),
		Error:         ErrUnexpectedType.Error(),
		MarshaledBack: `null`,
		JSONType:      Number,
		GoMapping:     GoInvalidMapping,
	}, //*/

	{
		Name:          "Bugfix: array and object factory overwrite each other",
		JSONData:      []byte(`{"some":"data"}`),
		Payload:       AcquirePayload().WithObject().WithArray(),
		Error:         "",
		MarshaledBack: `{"some":"data"}`,
		JSONType:      Object,
		GoMapping:     GoOther,
	}, //*/

	/* Template
	{
		Name:          "",
		JSONData:      []byte(``),
		Payload:       AcquirePayload(),
		Error:         "",
		MarshaledBack: ``,
		JSONType:      InvalidJSON,
		GoMapping:     GoInvalidMapping,
	}, //*/

}

func testPayloadHelper(test PayloadTest, reset bool, parallel bool,
) func(*testing.T) {
	return func(t *testing.T) {
		if parallel {
			t.Parallel()
		}

		// Run Unmarshaler.
		var strErr string
		err := test.Payload.UnmarshalJSON(
			test.JSONData)
		if err != nil {
			strErr = err.Error()
		}

		// Assert expected error.
		if (strErr != "" && test.Error == "") ||
			!strings.Contains(strErr, test.Error) {
			t.Fatalf("[%s] Unexpected marshal error\nWant Error: %s"+
				"\nHave Error: %s", test.Name, test.Error,
				strErr)
		}

		// Assert JSON Type.
		if test.JSONType != test.Payload.GetJSONType() {
			t.Fatalf("[%s] Unexpected JSON Data Type\nWant Type: %d"+
				"\nHave Type: %d", test.Name, test.JSONType,
				test.Payload.jsonType)
		}

		// Marshal back to JSON and assert no side effects in the data.
		d, _ := test.Payload.Get()
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("[%s] Unexpected error in marshal back: %v", test.Name,
				err)
		}
		if string(b) != test.MarshaledBack {
			t.Fatalf("[%s] Failed marshaling back\nWant MarshalBack:"+
				" %s\nHave MarshalBack: %s", test.Name,
				test.MarshaledBack, b)
		}

		// Assert the correctness of the reported GoMapping.
		wantMap := test.GoMapping
		haveMap := test.Payload.GetMapping()
		if haveMap != wantMap {
			t.Fatalf("[%s] Unexpected GoMapping. Want: %v; Have: %v",
				test.Name, wantMap, haveMap)
		}

		// Test that we are capable of correctly reset the Payload.
		if reset {
			test.Payload.Reset()
			p := fmt.Sprintf("%#v", test.Payload)
			if p != emptyPayload {
				t.Fatalf("[%s] Failed to fully reset Payload.\nWant: %v\n"+
					"Have: %v", test.Name, emptyPayload, p)
			}
			ReleasePayload(test.Payload)
		}

	}
}

func TestPayload_Raw(t *testing.T) {
	t.Parallel()
	for i := range TestsPayloadRaw {
		t.Run(TestsPayloadRaw[i].Name, testPayloadHelper(TestsPayloadRaw[i],
			true, true))
	}
}

func TestPayload_Getters(t *testing.T) {
	test := PayloadTest{Payload: AcquirePayload()}

	test.Name = "Getters test - Bool"
	test.Payload.WithBoolean()
	test.MarshaledBack = `true`
	test.JSONType = Boolean
	test.GoMapping = GoBool
	test.Error = ""
	test.JSONData = []byte(test.MarshaledBack)
	testPayloadHelper(test, false, false)(t)
	if test.Payload.GetBool() != true {
		t.Fatalf("Want: %v; Have: %v", true, false)
	}

	test.Payload.Reset()
	test.Name = "Getters test - Float"
	test.MarshaledBack = `3.14`
	test.Payload.WithFloat()
	test.JSONType = Number
	test.GoMapping = GoFloat
	test.JSONData = []byte(test.MarshaledBack)
	testPayloadHelper(test, false, false)(t)
	if val := test.Payload.GetFloat(); val != 3.14 {
		t.Fatalf("Want: %v; Have: %v", 3.14, val)
	}

	test.Payload.Reset()
	test.Name = "Getters test - Int"
	test.MarshaledBack = `314`
	test.Payload.WithInt()
	test.JSONType = Number
	test.GoMapping = GoInt
	test.JSONData = []byte(test.MarshaledBack)
	testPayloadHelper(test, false, false)(t)
	if val := test.Payload.GetInt(); val != 314 {
		t.Fatalf("Want: %v; Have: %v", 314, val)
	}

	test.Payload.Reset()
	test.Name = "Getters test - Null"
	test.MarshaledBack = `null`
	test.Payload.WithNull()
	test.JSONType = Null
	test.GoMapping = GoNil
	test.JSONData = []byte(test.MarshaledBack)
	testPayloadHelper(test, false, false)(t)
	if val := test.Payload.IsNil(); val != true {
		t.Fatalf("Want: %v; Have: %v", true, val)
	}

	test.Payload.Reset()
	test.Name = "Getters test - Uint"
	test.MarshaledBack = `314`
	test.Payload.WithUint()
	test.JSONType = Number
	test.GoMapping = GoUint
	test.JSONData = []byte(test.MarshaledBack)
	testPayloadHelper(test, false, false)(t)
	if val := test.Payload.GetUint(); val != 314 {
		t.Fatalf("Want: %v; Have: %v", 314, val)
	}

	var panicVal interface{}
	func() {
		defer func() {
			panicVal = recover()
		}()
		test.Payload.GetArray()
	}()
	if val := panicVal.(Error); val != ErrUnexpectedMapping {
		t.Fatalf("Expecting panic value to be Error ErrUnexpectedMapping."+
			" Got: %v", val)
	}

}

func TestAcquirePayload(t *testing.T) {
	p := AcquirePayload()
	ReleasePayload(p)
	p = AcquirePayload()
	ReleasePayload(p)
}
