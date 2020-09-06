# jsonutils [![PkgGoDev](https://pkg.go.dev/badge/github.com/diegommm/jsonutils?tab=doc)](https://pkg.go.dev/github.com/diegommm/jsonutils?tab=doc) [![GoDoc](https://godoc.org/github.com/diegommm/jsonutils?status.svg)](https://godoc.org/github.com/diegommm/jsonutils) [![Go Report Card](https://goreportcard.com/badge/github.com/diegommm/jsonutils)](https://goreportcard.com/report/github.com/diegommm/jsonutils) [![codecov](https://codecov.io/gh/diegommm/jsonutils/branch/master/graph/badge.svg)](https://codecov.io/gh/diegommm/jsonutils)

Utilities to handle JSON encoded data

## Determine JSON Data Type

Sometimes we receive JSON requests/responses from the internet and the API we're hitting (or clients hitting our API) always sends us a structure that could be modeled like this:
```go
type Response struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
}
```
And it can get a bit clunky to just leave it like a `map[string]interface{}` in case we get an object, or maybe a `[]interface{}` if it's an array or even just a plain `string`. We may want to validate the structure of what we're about to receive.

Let's say that we are making request to an API that lists objects of type `User`. The `User` model could look like this:
```go
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
```
Our query could return zero, one or more elements in the `Data` element depending on the matched criteria, returning JSON structs like the following:

Zero matches:
```json
{"status":200,"data":"no users matched the selected criteria"}
```
One match:
```json
{"status":200,"data":{"id":123,"username":"johndoe","email":"john@doe.com"}}
```
More than one match:
```json
{"status":200,"data":[{"id":123,"username":"johndoe","email":"john@doe.com"},{"id":321,"username":"jeanndoe","email":"jean@doe.com"}]}
```
The `TypeOf` function allows inspects the first bytes of a `[]byte` to try to guess the JSON Data Type held in it and returns a constant and an error. We can use this function modifying the initial `Response` struct like the following:
```go
type Response struct {
	Data   json.RawMessage `json:"data"`
	Status int             `json:"status"`
}
```
This type will instruct the JSON unmarshaler to store the raw bytes corresponding to that specific field instead of decoding them. Our decoding routine now can look something like this:
```go
// first unmarshaling round.
resp := new(Response)
if err := json.Unmarshal(theBytes, resp); err != nil {
    // handle error
}

// determine underlying data structure.
t, err := jsonutils.TypeOf(resp.Data)
if err != nil {
    // handle error
}

// last unmarshaling round: conditional unmarshal.
var users []User
switch t {
case jsonutils.String:
    fmt.Printf("Got message: %s\n", resp.Data)

case jsonutils.Array:
    if err := json.Unmarshal(resp.Data, &users); err != nil {
        // handle error
    }

case jsonutils.Object:
    resp = make([]User, 1)
    if err := json.Unmarshal(resp.Data, &users[0]); err != nil {
        // handle error
    }
}
```

## Variable-type payloads

TODO
