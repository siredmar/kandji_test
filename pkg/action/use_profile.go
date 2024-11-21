package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/service"
)

func UseProfile(s *service.Service, profile string) error {
	if err := s.Config.Write("CurrentProfile", profile); err != nil {
		return err
	}
	fmt.Printf("Switched to profile \"%s\"\n", s.Config.Auth.CurrentProfile)

	return nil
}
