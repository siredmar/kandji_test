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

func InvalidFileFormat() error {
	return errors.New("The provided file is not in a valid format")
}

func InvalidParameter(param string, hint string) error {
	s := fmt.Sprintf("The parameter %s is invalid. \n See '%s' for help and examples.", param, hint)
	return errors.New(s)
}

func MissingParameter(param string, hint string) error {
	s := fmt.Sprintf("The required parameter %s is missing. \n See '%s' for help and examples.", param, hint)
	return errors.New(s)
}

func NothingToDo(hint string) error {
	s := fmt.Sprintf("Nothing to do. \n See '%s' for help and examples.", hint)
	return errors.New(s)
}

func ServerError(msg string) error {
	s := fmt.Sprintf("Unexpected server error: %s", msg)
	return errors.New(s)
}
