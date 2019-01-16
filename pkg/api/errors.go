package api

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
		s = fmt.Sprintf("No %s found", d.Command)
	}
	fmt.Println(s)
	err := errors.New(s)
	return err
}

func NotImplementedError() error {
	s := fmt.Sprintf("This is not yet implemented ;(")
	fmt.Println(s)
	err := errors.New(s)
	return err
}
