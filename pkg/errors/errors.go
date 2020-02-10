// Package errors defines the error handling used by backend components.
package errors

import (
	"bytes"
	"strings"
)

// Error implements the the error interface.
// It contains a number of fields and an Error value may leave some values
// unset
type Error struct {
	// Err is the underlying error that triggered this one, if any.
	Err error

	// Kind classifies the kind of error, such as validation failed.
	// Other is returned if its class is unknown or not relevant.
	Kind Kind

	// Message represents the message reported to the user.
	Message string

	// Details represents detail information for the user to fix
	// this problem.
	Details []string
}

// Kind defines the kind of error.
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common.
// Do not reorder this list or remove any items since that it will change
// their values.
// New items muss be added only to the end.
const (
	Other          Kind = iota // Unclassified error
	Permission                 // Permission denied
	Invalid                    // Invalid action
	Validation                 // Validation failed
	Exist                      // Resource already exists
	NotExists                  // Resource does not exist
	Service                    // External service is not reachable
	NotImplemented             // Not yet implemented
	Internal                   // Internal error
)

// String returns the string representation of the error kind.
func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Permission:
		return "permission denied"
	case Invalid:
		return "invalid action"
	case Validation:
		return "validation failed"
	case Exist:
		return "resource already exists"
	case NotExists:
		return "resource does not exist"
	case Service:
		return "service unavailable"
	case NotImplemented:
		return "not implemented"
	case Internal:
		return "internal error"
	}
	return "unknown error kind"
}

// E returns an error based on the input.
// It constructs the error instance based on the value types given.
func E(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Kind:
			e.Kind = arg
		case *Error:
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		case string:
			e.Message = arg
		case []string:
			e.Details = arg
		}

	}

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// The previous error was also one of ours. Suppress duplications
	// so the message won't contain the same kind, file name or user name
	// twice.
	if prev.Kind == e.Kind {
		prev.Kind = Other
	}
	// If this error has Kind unset or Other, pull up the inner one.
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}

	return e
}

// Error implements the error interface.
func (e *Error) Error() string {
	b := new(bytes.Buffer)
	if e.Message != "" {
		b.WriteString(e.Message + " ")
	}
	if e.Kind != 0 {
		b.WriteString("(" + e.Kind.String() + ")")
	}
	if len(e.Details) != 0 {
		b.WriteString("\n" + strings.Join(e.Details, " "))
	}
	if e.Err != nil {
		b.WriteString("\n" + e.Err.Error())
	}
	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// IsKind reports whether a error is a specific kind.
func IsKind(e error, k Kind) bool {
	err, ok := e.(*Error)
	if !ok {
		return false
	}

	return err.Kind == k
}
