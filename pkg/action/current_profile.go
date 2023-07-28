package action

import (
	"fmt"
	"github.com/grid-x/gxctl/pkg/service"
)

func CurrentProfile(s *service.Service) error {
	currentProfile := s.Config.Auth.CurrentProfile
	if currentProfile == "" {
		return fmt.Errorf("No current profile set. Use gxctl config use-profile PROFILE_NAME to set it")
	} else {
		fmt.Println(currentProfile)
	}
	return nil
}
