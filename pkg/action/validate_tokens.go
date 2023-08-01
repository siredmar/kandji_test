package action

import (
	"fmt"
	"github.com/grid-x/gxctl/pkg/service"
	"github.com/grid-x/gxctl/pkg/spinner"
	"strings"
)

func ValidateTokens(s *service.Service) error {
	for i, p := range s.Client.Auth.Profiles {
		auth := spinner.New("")
		var sb strings.Builder
		err := p.Auth.Token.Validate()
		if err == nil {
			sb.WriteString(fmt.Sprintf("\u001B[32mOK  \u001B[39m Profile %s: \033[32mvalid token\033[39m", p.Name))
			auth.Write(sb.String())
		} else {
			sb.WriteString(fmt.Sprintf("\u001B[31mFAIL\u001B[39m Profile %s: \033[33m%s\033[39m", p.Name, err))
			auth.Write(sb.String())
		}

		if i < len(s.Client.Auth.Profiles)-1 {
			sb.WriteString("\n")
		}
	}

	return nil
}
