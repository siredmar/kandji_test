package errors

import (
	"errors"
	"fmt"
)

type ErrorDetails struct {
	Command, Id string
}

func NotFoundError(d ErrorDetails) error {
	var s string
	if d.Id != "" {
		s = fmt.Sprintf("Error from server (NotFound): %s \"%s\" not found", d.Command, d.Id)
	} else {
		s = fmt.Sprintf("No %s found\n", d.Command)
	}
	return errors.New(s)
}

func NotImplementedError(msg string) error {
	s := fmt.Sprintf("%s is not yet implemented ;(", msg)
	return errors.New(s)
}
