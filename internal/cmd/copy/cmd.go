package copy

import (
	"strings"

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
		Use:   "copy SOURCE DESTINATION",
		Short: "forwards files to devices",
		Run: func(cmd *clix.Command, args []string) error {
			if len(args) != 2 {
				return errors.E(
					errors.Invalid,
					"required argument ID not found",
					[]string{"run 'gxctl copy --help' for usage"},
				)
			}

			sourceID := args[0]
			destID := args[1]

			if strings.Contains(sourceID, ":") && strings.Contains(destID, ":") {
				return errors.E(
					errors.Invalid,
					"You cannot specify the device in both source and destination parameters",
					[]string{"run 'gxctl copy --help' for usage"},
				)
			}
			if !strings.Contains(sourceID, ":") && !strings.Contains(destID, ":") {
				return errors.E(
					errors.Invalid,
					"Missing device id",
					[]string{"run 'gxctl copy --help' for usage"},
				)
			}

			return action.Copy(s, sourceID, destID)
		},
	}

	return nil
}
