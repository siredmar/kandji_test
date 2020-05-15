package application

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/errors"
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
		Use:     "application NAME",
		Aliases: []string{"applications", "app", "apps"},
		Short:   "delete app",
		Run: func(cmd *clix.Command, args []string) error {
			if len(args) < 1 {
				return errors.E(
					errors.Invalid,
					"required argument NAME not found",
					[]string{"run 'gxctl delete application --help' for usage"},
				)
			}

			return action.DeleteApplication(s, args)
		},
	}

	return nil
}
