// Package errors defines the error handling used by backend components.
package errors

import (
	"bytes"
)

// Error implements the the error interface.
// It contains a number of fields and an Error value may leave some values
// unset
type Error struct {
	// Path is the URL path of requested API endpoint.
	Path PathName `json:"-"`
	// User is the id of the user attempting the action.
	User UserID `json:"-"`
	// Action is the action being performed, usually the name of the method and
	// package being invoked.
	Action Action `json:"-"`
	// Kind classifies the kind of error, such as validation failed.
	// Other is returned if its class is unknown or not relevant.
	Kind Kind `json:"-"`
	// Err is the underlying error that triggered this one, if any.
	Err error `json:"-"`

	// Message represents the message reported to the user.
	Message string `json:"message"`

	// Details represents detail information for the user to fix
	// this problem.
	Details []string `json:"details,omitempty"`
}

// PathName is just the url path representation.
type PathName string

// UserID is just the id representing the user.
type UserID string

// Action describes an action, usually as the package and method,
// such as "postgres/Account.Create".
type Action string

// Kind defines the kind of error.
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common.
// Do not reorder this list or remove any items since that it will change
// their values.
// New items muss be added only to the end.
const (
	Other      Kind = iota // Unclassified error
	Permission             // Permission denied
	Invalid                // Invalid action
	Validation             // Validation failed
	Exist                  // Resource already exists
	NotExists              // Resource does not exist
	Service                // External service is not reachable
	BadRequest             // Bad request
	Internal               // Internal error
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
	case BadRequest:
		return "bad request"
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
		case PathName:
			e.Path = arg
		case UserID:
			e.User = arg
		case Action:
			e.Action = arg
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
	if prev.Path == e.Path {
		prev.Path = ""
	}
	if prev.User == e.User {
		prev.User = ""
	}
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

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

// Error implements the error interface.
func (e *Error) Error() string {
	b := new(bytes.Buffer)
	if e.Action != "" {
		pad(b, ": ")
		b.WriteString(string(e.Action))
	}
	if e.Path != "" {
		pad(b, ": ")
		b.WriteString(string(e.Path))
	}
	if e.User != "" {
		pad(b, "; ")
		b.WriteString("user ")
		b.WriteString(string(e.Path))
	}
	if e.Kind != 0 {
		pad(b, ": ")
		b.WriteString(e.Kind.String())
	}
	if e.Err != nil {
		pad(b, ": ")
		b.WriteString(e.Err.Error())
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
