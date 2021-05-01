package deviceconfigmap

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
		Use:     "deviceconfigmap ID",
		Aliases: []string{"configmaps", "configmap", "cm", "deviceconfigmaps", "dcm"},
		Short:   "delete device config map",
		Args: args.Args{
			args.ValidateSingle("ID"),
			args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			return action.DeleteDeviceConfigMap(s, args)
		},
	}

	return nil
}
