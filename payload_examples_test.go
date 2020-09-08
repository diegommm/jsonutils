package jsonutils

import (
	"encoding/json"
	"log"
)

func ExamplePayload() {
	// This is intends to be a possible depiction of an API service.

	// Generic response type. Errors can be _fatal_ or _mon-fatal_. In the first
	// case, `data` will be set to `null` and `error` will not be present. In
	// the latter, `data` will be set to `null` and `error` will specify the
	// reason.
	//
	// A fatal error happens when the API Server cannot respond to a request due
	// to lack of privileges of the user, lack of authentication or other
	// similar errors.
	type Response struct {
		Status  int      `json:"status"`          // HTTP Status
		Payload *Payload `json:"data"`            // Expected, variable payload
		Error   *string  `json:"error,omitempty"` // Optional error string
	}

	// The GET Users endpoint returns either an array of users, a single user
	// object or a string explaining why no user was found.
	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// This could be the expected payload for a specific endpoint.
	GetUser := Response{Payload: AcquirePayload()}
	defer ReleasePayload(GetUser.Payload)

	// Upon success, the endpoint returns a User object.
	GetUser.Payload.WithObject(func() interface{} {
		return new(User)
	})

	// It can also be a string stating solely the name of the user. E.g.: the
	// endpoint might receive a Query String Param specifying this.
	GetUser.Payload.WithString()

	// If we get an error then Data will be set to nil.
	GetUser.Payload.WithNull()

	respPreamble, respEpilogue := `{"status":200,"data":`, `}`

	// Exercise some payloads
	payloads := map[string][]byte{
		"success": []byte(respPreamble + `{"id":120,"username":"diegommm",` +
			`"email":"diegoaugustomolina@gmail.com"}` + respEpilogue),
		"string":     []byte(respPreamble + `"The user is diegommm"` + respEpilogue),
		"error":      []byte(respPreamble + `null` + respEpilogue),
		"unexpected": []byte(respPreamble + `1` + respEpilogue),
	}

	// Successful response example.
	if err := json.Unmarshal(payloads["success"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (success): %v", err)
	}
	if _, ok := GetUser.Payload.GetObject().(*User); !ok {
		log.Fatalf("unexpected decoded data from JSON (want: User): %v",
			GetUser.Payload)
	}

	// Alternative successful response example, with string.
	if err := json.Unmarshal(payloads["string"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (string): %v", err)
	}
	s := respPreamble + "\"" + GetUser.Payload.GetString() + "\"" + respEpilogue
	if s != string(payloads["string"]) {
		log.Fatalf("unexpected decoded data from JSON: Want: %s; Got: %s",
			payloads["string"], s)
	}

	// Unsuccessful response example, receive null.
	if err := json.Unmarshal(payloads["error"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (null): %v", err)
	}
	// As here we are using a pointer to Payload inside another struct then the
	// JSON Unmarshaler will set this pointer to nil. If we were not using a
	// pointer then we could use the IsNil method.
	if GetUser.Payload != nil {
		log.Fatalf("unexpected decoded data from JSON (want: nil): %#v",
			GetUser.Payload)
	}

	// Unexpected data type: Number
	err := json.Unmarshal(payloads["unexpected"], &GetUser)
	if err != ErrUnexpectedType {
		log.Fatalf("unexpected error decoding JSON (want: ErrUnexpectedType)"+
			": %v", err)
	}
	if GetUser.Payload.GetJSONType() != Number {
		log.Fatalf("unexpected decoded data from JSON (want: Numbaer): %#v",
			GetUser.Payload)
	}

	// Output:
	//
}
