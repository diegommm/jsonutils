package jsonutils

import "testing"

func TestUtils(t *testing.T) {
	var m GoMapping
	var panicVal interface{}

	func() {
		defer func() {
			panicVal = recover()
		}()
		m.panicIfNot(GoOther)
	}()

	if panicVal == nil {
		t.Fatalf("Expected panicVal to be non-nil")
	}

	e, ok := panicVal.(Error)
	if !ok {
		t.Fatalf("Expected panicVal to be of type Error. Got: %#v", panicVal)
	}

	if e != ErrUnexpectedMapping {
		t.Fatalf("Expected panicVal to contain ErrUnexpectedMapping. Got: %#v",
			e)
	}

	var dummy noCopy
	dummy.Lock()
	dummy.Unlock()
}
