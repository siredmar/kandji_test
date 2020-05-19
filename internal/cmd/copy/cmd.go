package copy

import (
	"strings"

	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
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
		Args: args.Args{
			validateArgs(),
			args.PredictFile(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			sourceID := args[0]
			destID := args[1]
			return action.Copy(s, sourceID, destID)
		},
	}

	return nil
}

func validateArgs() clix.ValidateFunc {
	return func(args []string) error {
		if len(args) != 2 {
			return errors.E(
				errors.Invalid,
				"required arguments source and dest not found",
				nil,
			)
		}

		sourceID := args[0]
		destID := args[1]

		if strings.Contains(sourceID, ":") && strings.Contains(destID, ":") {
			return errors.E(
				errors.Invalid,
				"You cannot specify the device in both source and destination parameters",
				nil,
			)
		}
		if !strings.Contains(sourceID, ":") && !strings.Contains(destID, ":") {
			return errors.E(
				errors.Invalid,
				"Missing device id",
				nil,
			)
		}
		return nil
	}
}
