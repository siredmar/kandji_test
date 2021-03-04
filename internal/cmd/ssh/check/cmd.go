package check

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/service"
	"github.com/grid-x/gxctl/pkg/errors"
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
		Use:   "check",
		Short: "Check if current SSH config is valid",
		Run: func(cmd *clix.Command, args []string) error {
			err := action.SSHConfigCheck(s)
			if err != nil {
				return errors.E(err, "SSH config differs from expected")
			}
			return nil
		},
	}

	return nil
}
