package config

import (
	"errors"
	"fmt"
)

var ErrNoKey = errors.New("config key does not exist")
var ErrNilValue = errors.New("config key has no or invalid value")
var ErrInvalidValue = errors.New("value is invalid")

// -- //
type ConfigError struct {
	Key   string
	Value string
	Err   error
}

func (e ConfigError) Error() (err string) {
	switch {
	case e.Key != "" && e.Value != "":
		err = fmt.Sprintf("key=%s value=%s: %s", e.Key, e.Value, e.Err.Error())
	case e.Key != "":
		err = fmt.Sprintf("key=%s value=<nil>: %s", e.Key, e.Err.Error())
	}

	return
}
