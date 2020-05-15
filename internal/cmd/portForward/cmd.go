package portForward

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
		Use:   "forward",
		Short: "forwards remote connections to local port",
		Run: func(cmd *clix.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument ID not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
			}

			portForwardCmdLocalPort, _ := cmd.Flags().GetString("localport")
			portForwardCmdTarget, _ := cmd.Flags().GetString("target")

			if portForwardCmdLocalPort == "" {
				return errors.E(
					errors.Invalid,
					"required parameter '--localport' not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
			}
			if portForwardCmdTarget == "" {
				return errors.E(
					errors.Invalid,
					"required parameter '--target' not found",
					[]string{"run 'gxctl port-forward --help' for usage"},
				)
			}

			return action.PortForward(s, args[0], portForwardCmdLocalPort, portForwardCmdTarget)
		},
	}

	c.cmd.Flags().StringP("localport", "l", "", "Local port for the forwarded connection")
	c.cmd.Flags().StringP("target", "t", "", "Target to forward traffic from")

	return nil
}
