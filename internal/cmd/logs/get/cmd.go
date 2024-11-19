package get

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/service"
)

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

func (c *CMD) Init(s *service.Service) error {
	c.cmd = &clix.Command{
		Use:   "get DEVICE_ID (or serial number if used with the -S flag)",
		Short: "get the current state of the logs settings for a given device",
		Args: args.Args{
			Validator: args.ValidateSingle("device ID"),
		},
		Run: func(cmd *clix.Command, args []string) error {
			output, _ := c.cmd.Flags().GetString("output")
			isSerialNumber, _ := c.cmd.Flags().GetBool("serial")

			return action.GetLogs(s, args[0], isSerialNumber, output)
		},
	}

	c.cmd.Flags().StringP("output", "o", "", "Print result in a different format. Must be one of: json|wide|yaml")
	c.cmd.Flags().BoolP("serial", "S", false, "treat device ID as a serial number")

	return nil
}
