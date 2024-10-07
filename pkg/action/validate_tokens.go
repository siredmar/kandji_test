package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/service"
)

const (
	redGreenClosingTag = "\u001B[39m"
	redTag             = "\u001B[31m"
	greenTag           = "\u001B[32m"
	yellowTag          = "\033[33m"
	closingYellowTag   = "\033[0m"
	ok                 = "OK  "
	fail               = "FAIL"
	validToken         = "valid token"
)

func ValidateTokens(s *service.Service) error {
	fmt.Printf("CurrentProfile: %q\n", s.Client.Auth.CurrentProfile)

	for _, p := range s.Client.Auth.Profiles {
		err := p.Auth.Token.Validate()
		if err == nil {
			fmt.Printf("%s Profile %q: %s\n", printGreen(ok), p.Name, printYellow(validToken))
		} else {
			fmt.Printf("%s Profile %q: %s\n", printRed(fail), p.Name, printYellow(err.Error()))
		}
	}

	return nil
}

func printRed(s string) string {
	return redTag + s + redGreenClosingTag
}

func printGreen(s string) string {
	return greenTag + s + redGreenClosingTag
}

func printYellow(s string) string {
	return yellowTag + s + closingYellowTag
}
