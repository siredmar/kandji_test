package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/service"
)

func ValidateTokens(s *service.Service) error {
	for _, p := range s.Client.Auth.Profiles {
		err := p.Auth.Token.Validate()
		if err == nil {
			fmt.Printf("\u001B[32mOK  \u001B[39m Profile %s: \033[32mvalid token\033[0m\n", p.Name)
		} else {
			fmt.Printf("\u001B[31mFAIL\u001B[39m Profile %s: \033[33m%s\033[0m\n", p.Name, err)
		}
	}

	return nil
}
