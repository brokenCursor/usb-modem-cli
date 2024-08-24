package drivers

import (
	"errors"
	"fmt"
)

// Basic Errors
var ErrUnknownModel = errors.New("attempting to get unknown model")
var ErrNoDrivers = errors.New("no drivers were registered")
var ErrUnknown = errors.New("unknown error")

// Complex Errors
type ActionError struct {
	Action string
	Err    error
}

func (e ActionError) Unwrap() error {
	return e.Err
}

func (e ActionError) Error() string {
	return fmt.Sprintf("action error: %s failed with %e", e.Action, e.Err)
}

// -- //
type UnmarshalError struct {
	RawData *[]byte
	Err     error
}

func (e UnmarshalError) Unwrap() error {
	return e.Err
}

func (e UnmarshalError) Error() string {
	return "failed to unmarshal response"
}
