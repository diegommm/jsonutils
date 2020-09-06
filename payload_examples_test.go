package jsonutils

import (
	"encoding/json"
	"log"
)

func ExamplePayload() {
	// Generic response type. To be reused.
	type Response struct {
		Status  int      `json:"status"`
		Payload *Payload `json:"data"`
		Error   *string  `json:"error,omitempty"`
	}

	// Specific structure of one of the possible responses
	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// This could be the expected payload for a specific endpoint.
	GetUser := Response{Payload: new(Payload)}

	// Upon success, the endpoint returns a User object.
	GetUser.Payload.WithObject(func() interface{} {
		return new(User)
	})

	// It can also be a string stating solely the name of the user. E.g.: the
	// endpoint might receive a Query String Param specifying this.
	GetUser.Payload.WithString()

	// If we get an error then Data will be set to nil.
	GetUser.Payload.WithNull()

	respPreamble, repEpilogue := `{"status":200,"data":`, `}`

	// Excercise some payloads
	payloads := map[string][]byte{
		"success": []byte(respPreamble + `{"id":120,"username":"diegommm",` +
			`"email":"diegoaugustomolina@gmail.com"}` + repEpilogue),
		"string":     []byte(respPreamble + `"The user is diegommm"` + repEpilogue),
		"error":      []byte(respPreamble + `null` + repEpilogue),
		"unexpected": []byte(respPreamble + `1` + repEpilogue),
	}

	// Successful response example.
	if err := json.Unmarshal(payloads["success"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (success): %v", err)
	}
	if _, ok := GetUser.Payload.Data.(*User); !ok {
		log.Fatalf("unexpected decoded data from JSON (want: User): (%T) %v",
			GetUser.Payload.Data, GetUser.Payload.Data)
	}

	// Alternative successful response example, with string.
	if err := json.Unmarshal(payloads["string"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (string): %v", err)
	}
	if _, ok := GetUser.Payload.Data.(*string); !ok {
		log.Fatalf("unexpected decoded data from JSON (want: string): (%T) %v",
			GetUser.Payload.Data, GetUser.Payload.Data)
	}

	// Unsuccessful response example, receive to null.
	if err := json.Unmarshal(payloads["error"], &GetUser); err != nil {
		log.Fatalf("unexpected error decoding JSON (null): %v", err)
	}
	if GetUser.Payload != nil {
		log.Fatalf("unexpected decoded data from JSON (want: nil): (%T) %v",
			GetUser.Payload.Data, GetUser.Payload.Data)
	}

	// Unexpected data type: Number
	err := json.Unmarshal(payloads["unexpected"], &GetUser)
	if err != ErrUnexpectedType {
		log.Fatalf("unexpected error decoding JSON (want: ErrUnexpectedType)"+
			": %v", err)
	}
	if GetUser.Payload.Data != nil {
		log.Fatalf("unexpected decoded data from JSON (want: nil): (%T) %v",
			GetUser.Payload.Data, GetUser.Payload.Data)
	}

	// Output:
	//
}
