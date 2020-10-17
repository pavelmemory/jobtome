package internal

import (
	"errors"
)

// ErrBadInput signals that one of the parameters/arguments for the method/function
// has a bad value: empty, too big, too small, etc.
var ErrBadInput = errors.New("bad input")

// ErrNotUnique shows that the value used already exists (is a duplicate of another value).
var ErrNotUnique = errors.New("not unique")

// ErrNotFound shows that the requested value was not found (or doesn't exist).
var ErrNotFound = errors.New("not found")
