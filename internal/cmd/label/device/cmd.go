package device

import (
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
		Use:     "device ID KEY_1=VAL_1 ... KEY_N=VAL_N",
		Aliases: []string{"devices"},
		Short:   "label device",
		Args: args.Args{
			validateArgs(),
			args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			return action.LabelDevice(s, args)
		},
	}

	return nil
}

func validateArgs() clix.ValidateFunc {
	return func(args []string) error {
		if len(args) == 0 {
			return errors.E(
				errors.Invalid,
				"required argument ID not found",
				[]string{"run 'gxctl label device --help' for usage"},
			)
		}

		if len(args) == 1 {
			return errors.E(
				errors.Invalid,
				"required arguments KEY=VAL not found",
				[]string{"run 'gxctl label device --help' for usage"},
			)
		}
		return nil
	}
}
