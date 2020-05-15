package get

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/internal/cmd/get/application"
	"github.com/grid-x/gxctl/internal/cmd/get/deployment"
	"github.com/grid-x/gxctl/internal/cmd/get/device"
	"github.com/grid-x/gxctl/internal/cmd/get/maintenance"
	"github.com/grid-x/gxctl/internal/cmd/get/pod"
	"github.com/grid-x/gxctl/pkg/service"
)

// CMD contains a command and all its sub commands
type CMD struct {
	cmd      *clix.Command
	children []cmd.CMD
}

// New returns a new CMD
func New() *CMD {
	return &CMD{}
}

// Command returns the internal *clix.Command
func (c *CMD) Command() *clix.Command {
	return c.cmd
}

// Children returns subcommands
func (c *CMD) Children() []cmd.CMD {
	return c.children
}

// Init the Command
func (c *CMD) Init(s *service.Service) error {
	c.cmd = &clix.Command{
		Use:   "get",
		Short: "get different resources",
	}
	c.children = []cmd.CMD{
		application.New(),
		deployment.New(),
		device.New(),
		maintenance.New(),
		pod.New(),
	}

	return nil
}
