package update

import (
	"time"

	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/service"
)

var defaultDuration = time.Hour * 168 // 1 week

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
		Use:   "update deviceID (or serial number if used with the -S flag)",
		Short: "update logs settings for a given device",
		Args: args.Args{
			Validator: args.ValidateSingle("deviceID"),
		},
		Run: func(cmd *clix.Command, args []string) error {
			level, _ := cmd.Flags().GetString("level")
			isSerialNumber, _ := cmd.Flags().GetBool("serial")
			expiry, _ := cmd.Flags().GetDuration("expiry")
			output, _ := c.cmd.Flags().GetString("output")
			changeOwner, _ := c.cmd.Flags().GetBool("change-owner")

			return action.UpdateLogs(s, args[0], isSerialNumber, changeOwner, level, output, expiry)
		},
	}

	c.cmd.Flags().StringP("level", "l", "debug", "the desired log level for the device")
	c.cmd.Flags().BoolP("serial", "S", false, "treat device ID as a serial number")
	c.cmd.Flags().BoolP("change-owner", "c", false, "replace the current owner of the device logs settings with yourself. The new owner will receive notifications about logs settings for this device via slack")
	c.cmd.Flags().DurationP("expiry", "e", defaultDuration, "the expiry of these log settings depends on the duration specified. Example: 25h4m")
	c.cmd.Flags().StringP("output", "o", "yaml", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
