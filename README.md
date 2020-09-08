# jsonutils [![PkgGoDev](https://pkg.go.dev/badge/github.com/diegommm/jsonutils?tab=doc)](https://pkg.go.dev/github.com/diegommm/jsonutils?tab=doc) [![GoDoc](https://godoc.org/github.com/diegommm/jsonutils?status.svg)](https://godoc.org/github.com/diegommm/jsonutils) [![Go Report Card](https://goreportcard.com/badge/github.com/diegommm/jsonutils)](https://goreportcard.com/report/github.com/diegommm/jsonutils) [![codecov](https://codecov.io/gh/diegommm/jsonutils/branch/master/graph/badge.svg)](https://codecov.io/gh/diegommm/jsonutils)

Utilities to handle JSON encoded data

## Conditionally unmarshal

Sometimes we receive JSON requests/responses and the API we're hitting (or clients hitting our API) use a general structure that could be modeled like this:

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

Our query could returns may hold a varying structured `data` depending on the number of matches:

* Zero matches: `{"status":200,"data":"no users matched the selected criteria"}`
* One match:
`{"status":200,"data":{"id":123,"username":"johndoe","email":"john@doe.com"}}`
* More than one match: `{"status":200,"data":[{"id":123,"username":"johndoe","email":"john@doe.com"},{"id":321,"username":"jeandoe","email":"jean@doe.com"}]}`

The `TypeOf` function inspects the first bytes of a `[]byte` to guess the JSON Data Type held in it and returns a constant and an error. We can use this function and use a modified `Response` like the following:
```go
type Response struct {
	Data   json.RawMessage `json:"data"`
	Status int             `json:"status"`
}
```
The `json.RawMessage` will instruct the JSON standard unmarshaler to store the raw bytes corresponding to that specific field instead of decoding them. Our decoding routine now can look something like this:
<details><summary><b>Click to see the code!</b></summary>

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
</details>

The `json.Unmarshal` will correctly parse the value in the `data` element but will not perform any unmarshaling, avoiding expensive reflection calls until we know how to deal with the contents. The `TypeOf` function is very accurate and it needs to see no more than the first few of bytes.

## _The_ Conditional Unmarshaler

`Payload` allows you to do all this work for you very efficiently. You just need to create it, configure it to know what _should_ be unmarshaled and the you can get your data.

Consider the following model:
```go
type Response struct {
	Data   *jsonutils.Payload `json:"data"`
	Status int                `json:"status"`
}

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}
```
We will stick to the previous example where we could get a JSON Array, Object or String to create our API client:
<details><summary><b>Click to see the code!</b></summary>
	
```go
func GetUsers() ([]User, error) {
	var b [] byte
	// Make your API call and get data into b.

	// Create your response.
	resp := &Response{
		Data: jsonutils.AcquirePayload().
			// Allow JSON String.
			WithString().
			// Allow a single object with our User structure to be returned.
			WithObject(func() interface{} { return new(User) }).
			// Allow an array to be returned, saving it to a slice of User.
			// Note that we are returning a pointer to a nil slice.
			WithArray(func() interface{} { return new([]User) })
	}
	// This is not mandatory but recommended to improve performance.
	defer jsonutils.ReleasePayload(resp.Data)
	
	// We didn't configure our Payload to receive, say, a null value or an integer
	// so if we do, we will receive an unmarshaling error. To allow other types of
	// payloads see the docs (it's as easy as above!).
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}
	
	// See what we got for christmas.
	switch resp.Data.GetJSONType() {
	case jsonutils.Object:
		return []User{ *resp.Data.GetObject().(*User) }, nil
	case jsonutils.Array:
		return *resp.Data.GetArray().(*[]User), nil
	case jsonutils.String:
		return nil, fmt.Errorf("no results: %v", resp.Data.GetString())
	}
	
	panic("this won't happen")
}
```
</details>
There you go! Now you can reuse your structure and you don't have to write custom unmarshalers or, even worse, guess by unmarshaling iteratevly until you hit your expected structure.

### Limitations

`Payload` can only be set through unmarshaling. If you are in need for this feature, please, let me know. It shouldn't be that hard to add it but I didn't have the need to implement it yet.
