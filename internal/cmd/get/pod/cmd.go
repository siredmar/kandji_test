package pod

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
		Use:     "pod [NAME]",
		Aliases: []string{"pods", "po"},
		Short:   "get pod",
		Long:    "Print a list of all pods you have access to",
		Run: func(cmd *clix.Command, args []string) error {
			deviceID, _ := cmd.Flags().GetString("device-id")
			outputType, _ := cmd.Flags().GetString("output")
			showAll, _ := cmd.Flags().GetBool("all")
			sortBy, _ := cmd.Flags().GetString("sort-by")

			return action.GetPod(s, deviceID, outputType, sortBy, showAll, args)
		},
		Predictors: args.Predictors{
			"output":    args.PredictOutputType(),
			"device-id": args.PredictNil(),
			"sort-by":   args.PredictNil(),
		},
	}

	c.cmd.Flags().BoolP("all", "", false, "show also unstarted pods")
	c.cmd.Flags().String("output", "wide", "output format of result")
	c.cmd.Flags().StringP("device-id", "d", "", "specify device id")
	c.cmd.Flags().StringP("sort-by", "", "", "sort by")

	return nil
}
