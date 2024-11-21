package maintenance

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
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
		Use:     "maintenance ID",
		Aliases: []string{"maintenances"},
		Short:   "delete maintenance",
		Args: args.Args{
			Validator: args.ValidateSingle("ID"),
			Predictor: args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			return action.DeleteMaintenance(s, args)
		},
	}

	return nil
}
