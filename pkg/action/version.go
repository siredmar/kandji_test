package action

import (
	"github.com/grid-x/gxctl/internal/version"
	"github.com/grid-x/gxctl/pkg/service"
)

// Version prints the version info
func Version(s *service.Service) error {
	return version.VersionDetails()
}
