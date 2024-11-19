package disable

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
		Use:   "disable DEVICE_ID (or serial number if used with the -S flag)",
		Short: "disable logs settings for a given device",
		Args: args.Args{
			Validator: args.ValidateSingle("device ID"),
		},
		Run: func(cmd *clix.Command, args []string) error {
			isSerialNumber, _ := cmd.Flags().GetBool("serial")
			return action.DisableLogs(s, args[0], isSerialNumber)
		},
	}

	c.cmd.Flags().BoolP("serial", "S", false, "treat device ID as a serial number")

	return nil
}
