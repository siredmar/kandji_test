package sshnext

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/internal/cmd/sshnext/tunnel"
	"github.com/grid-x/gxctl/internal/cmd/sshnext/setup"
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
		Use:   "sshnext",
		Short: "ssh to different devices (experimental)",
	}
	c.children = []cmd.CMD{
		setup.New(),
		tunnel.New(),
	}


	return nil
}
