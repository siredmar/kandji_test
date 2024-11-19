package enable

import (
	"time"

	clix "github.com/go-clix/cli"

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
		Use:   "enable DEVICE_ID (or serial number if used with the -S flag)",
		Short: "enable logs setting for a given device",
		Run: func(cmd *clix.Command, args []string) error {
			level, _ := cmd.Flags().GetString("level")
			isSerialNumber, _ := cmd.Flags().GetBool("serial")
			duration, _ := cmd.Flags().GetDuration("expiration")
			output, _ := c.cmd.Flags().GetString("output")

			return action.EnableLogs(s, args[0], isSerialNumber, level, output, duration)
		},
	}

	c.cmd.Flags().StringP("level", "l", "debug", "the desired log level for the device")
	c.cmd.Flags().BoolP("serial", "S", false, "treat device ID as a serial number")
	c.cmd.Flags().DurationP("duration", "d", defaultDuration, "how long these logs settings are going to work. The value is parsed by time.ParseDuration from the go standard library")
	c.cmd.Flags().StringP("output", "o", "", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
