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
		Use:     "device [ID…]",
		Aliases: []string{"devices"},
		Short:   "get device",
		Long:    "print a list of all devices you have access to",
		Run: func(cmd *clix.Command, args []string) error {
			showDockerConfig, _ := cmd.Flags().GetBool("show-dockerconfig")
			showPods, _ := cmd.Flags().GetBool("show-pods")
			showPublicKey, _ := cmd.Flags().GetBool("show-publickey")

			if showDockerConfig && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"docker-configs can just be shown for a single device",
					[]string{"run 'gxctl get device --help' for usage"},
				)
			}

			if showPods && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"pods can just be shown for a single device",
					[]string{"run 'gxctl get device --help' for usage"},
				)
			}

			if showPublicKey && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"publickeys can just be shown for a single device",
					[]string{"run 'gxctl get device --help' for usage"},
				)
			}

			outputType, _ := cmd.Flags().GetString("output")
			showAll, _ := cmd.Flags().GetBool("all")
			sortBy, _ := cmd.Flags().GetString("sort-by")

			if err := action.GetDevice(s, outputType, sortBy, showDockerConfig, showPods, showPublicKey, showAll, args); err != nil {
				return err
			}

			return nil
		},
	}

	c.cmd.Flags().BoolP("all", "", false, "show also inactive devices")
	c.cmd.Flags().BoolP("show-dockerconfig", "", false, "print the docker config for a device")
	c.cmd.Flags().BoolP("show-pods", "", false, "print the pods for a device")
	c.cmd.Flags().BoolP("show-publickey", "", false, "print the public key for a device")
	c.cmd.Flags().StringP("sort-by", "", "", "sort by")
	c.cmd.Flags().String("output", "wide", "output format of result")

	return nil
}
