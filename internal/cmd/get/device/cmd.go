package device

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
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
			showDeployments, _ := cmd.Flags().GetBool("show-deployments")
			showDeploys, _ := cmd.Flags().GetBool("show-deploys")

			showDockerConfig, _ := cmd.Flags().GetBool("show-dockerconfig")
			showPods, _ := cmd.Flags().GetBool("show-pods")
			showPublicKey, _ := cmd.Flags().GetBool("show-publickey")

			if (showDeployments || showDeploys) && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"deployments can just be shown for a single device",
					nil,
				)
			}

			if showDockerConfig && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"docker-configs can just be shown for a single device",
					nil,
				)
			}

			if showPods && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"pods can just be shown for a single device",
					nil,
				)
			}

			if showPublicKey && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"publickeys can just be shown for a single device",
					nil,
				)
			}

			outputType, _ := cmd.Flags().GetString("output")
			showAll, _ := cmd.Flags().GetBool("all")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			label, _ := cmd.Flags().GetString("label")
			showPublicIP, _ := cmd.Flags().GetBool("show-public-ip")

			if err := action.GetDevice(s, outputType, label, sortBy, (showDeployments || showDeploys), showDockerConfig, showPods, showPublicIP, showPublicKey, showAll, args); err != nil {
				return err
			}

			return nil
		},
		Predictors: args.Predictors{
			"sort-by": args.PredictNil(),
			"output":  args.PredictOutputType(),
		},
	}

	c.cmd.Flags().Bool("all", false, "show also inactive devices")
	c.cmd.Flags().StringP("label", "l", "", "filter results by label")
	c.cmd.Flags().Bool("show-deployments", false, "print the deployments of a device")
	c.cmd.Flags().Bool("show-deploys", false, "print the deployments of a device")
	c.cmd.Flags().MarkHidden("show-deploys")
	c.cmd.Flags().Bool("show-dockerconfig", false, "print the docker config for a device")
	c.cmd.Flags().Bool("show-pods", false, "print the pods for a device")
	c.cmd.Flags().Bool("show-public-ip", false, "print the public ip for a device")
	c.cmd.Flags().Bool("show-publickey", false, "print the public key for a device")
	c.cmd.Flags().StringP("sort-by", "s", "", "sort by")
	c.cmd.Flags().StringP("output", "o", "", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
