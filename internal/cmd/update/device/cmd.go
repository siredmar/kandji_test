package device

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
		Use:   "device ID",
		Short: "update device",
		Run: func(cmd *clix.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument ID not valid",
					[]string{"run 'gxctl update device --help' for usage"},
				)
			}

			updateDeviceCmdMaintenanceWindow, _ := cmd.Flags().GetString("maintenance-window")
			updateDeviceCmdMacAddress, _ := cmd.Flags().GetString("mac-address")
			updateDeviceCmdLabels, _ := cmd.Flags().GetString("labels")
			updateDeviceCmdAnnotations, _ := cmd.Flags().GetString("annotations")

			if updateDeviceCmdMaintenanceWindow == "" && updateDeviceCmdMacAddress == "" && updateDeviceCmdLabels == "" && updateDeviceCmdAnnotations == "" {
				return errors.E(
					errors.Invalid,
					"Nothing to update",
					[]string{"run 'gxctl update device --help' for usage"},
				)
			}
			return action.UpdateDevice(
				s,
				args[0],
				updateDeviceCmdMaintenanceWindow,
				updateDeviceCmdMacAddress,
				updateDeviceCmdLabels,
				updateDeviceCmdAnnotations,
			)
		},
	}

	c.cmd.Flags().StringP("maintenance-window", "w", "", "Maintenance window for the device")
	c.cmd.Flags().StringP("mac-address", "m", "", "Mac address for the device")
	c.cmd.Flags().StringP("labels", "l", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")
	c.cmd.Flags().StringP("annotations", "a", "", "A space seperated list of annotations eg. gridx.ai/custimer=123 gridx.ai/style=red")

	return nil
}
