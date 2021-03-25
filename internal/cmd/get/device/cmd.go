package device

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
		Use:     "device [ID…]",
		Aliases: []string{"devices"},
		Short:   "get device",
		Long:    "print a list of all devices you have access to",
		Run: func(cmd *clix.Command, args []string) error {
			outputType, _ := cmd.Flags().GetString("output")
			showAll, _ := cmd.Flags().GetBool("all")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			label, _ := cmd.Flags().GetString("label")
			serial, _ := cmd.Flags().GetString("serial")
			showPublicIP, _ := cmd.Flags().GetBool("show-public-ip")
			showPublicKey, _ := cmd.Flags().GetBool("show-publickey")
			deploymentID, _ := cmd.Flags().GetString("deployment-id")

			if err := action.GetDevice(s, outputType, label, serial, sortBy, showPublicIP, showPublicKey, deploymentID, showAll, args); err != nil {
				return err
			}

			return nil
		},
		Predictors: args.Predictors{
			"sort-by":       args.PredictNil(),
			"deployment-id": args.PredictNil(),
			"output":        args.PredictOutputType(),
		},
	}

	c.cmd.Flags().Bool("all", false, "show also inactive devices")
	c.cmd.Flags().StringP("label", "l", "", "filter results by label")
	c.cmd.Flags().StringP("serial", "S", "", "filter results by serialnumber")
	c.cmd.Flags().Bool("show-public-ip", false, "print the public ip for a device")
	c.cmd.Flags().Bool("show-publickey", false, "print the public key for a device")
	c.cmd.Flags().StringP("deployment-id", "d", "", "specify deployment id")
	c.cmd.Flags().StringP("sort-by", "s", "", "sort by")
	c.cmd.Flags().StringP("output", "o", "", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
