package errors

import (
	"errors"
	"fmt"
)

func GetNotFoundError(id string) error {
	s := fmt.Sprintf("Error from server (NotFound): %s not found", id)
	return errors.New(s)
}

func ListNotFoundError(res string) error {
	s := fmt.Sprintf("No %s found", res)
	return errors.New(s)
}

func NotImplementedError(msg string) error {
	s := fmt.Sprintf("%s is not yet implemented ;(", msg)
	return errors.New(s)
}

func InvalidFormat() error {
	return errors.New("The provided file is not in a valid format")
}

func ServerError(msg string) error {
	s := fmt.Sprintf("Unexpected server error: %s", msg)
	return errors.New(s)
}
