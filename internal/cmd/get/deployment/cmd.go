package deployment

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
		Use:     "deployment [ID…]",
		Aliases: []string{"deployments", "deploy"},
		Short:   "get deployment",
		Long:    "print a list of all deployments you have access to",
		Run: func(cmd *clix.Command, args []string) error {
			outputType, _ := cmd.Flags().GetString("output")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			showDevices, _ := cmd.Flags().GetBool("show-devices")

			if showDevices && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"devices can just be shown for a single deployment",
					[]string{"run 'gxctl get deployment --help' for usage"},
				)
			}

			return action.GetDeployment(s, outputType, sortBy, showDevices, args)
		},
	}

	c.cmd.Flags().String("output", "wide", "output format of result")
	c.cmd.Flags().BoolP("show-devices", "", false, "print the devices for a deployment")
	c.cmd.Flags().StringP("sort-by", "", "", "sort by")

	return nil
}
