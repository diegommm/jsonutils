package jsonutils

import (
	"encoding/json"
	"fmt"
	"log"
)

func ExampleTypeOf() {
	// This is the general response form.
	type MyResponseType struct {
		Status int             `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	// This is one of the possible response payloads.
	type ResponsePayload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	// Suppose we receive this payload from a request.
	receivedPayload := []byte(
		`{"status":200,"data":{"key":"some key","value":"some value"}}`)

	// We first decode the general struct.
	r := MyResponseType{}
	err := json.Unmarshal(receivedPayload, &r)
	if err != nil {
		log.Fatal(err)
	}

	// We will assign some data type here according to the detected data type.
	var res interface{}

	// Then get the data type held by the variable-typed payload and test it.
	switch t, err := TypeOf(r.Data); {
	case err != nil:
		log.Fatal(err)
	case t == Object:
		log.Println("detected object payload")
		res = new(ResponsePayload)
	case t == Array:
		log.Println("detected array payload")
		res = new([]ResponsePayload)
	case t == String:
		log.Println("detected string payload")
		res = new(string)
	default:
		log.Fatalf("unexpected response: %s", r.Data)
	}

	// Finally, unmarshal into the corresponding data type.
	err = json.Unmarshal(r.Data, res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v", res)

	// Output:
	// &jsonutils.ResponsePayload{Key:"some key", Value:"some value"}
}
