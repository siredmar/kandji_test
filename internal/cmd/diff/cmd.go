package diff

import (
	"fmt"

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
		Use:   "diff",
		Short: "diff resources",
		Run: func(cmd *clix.Command, args []string) error {
			diffCmdFilename, _ := cmd.Flags().GetString("filename")
			diffCmdDiffer, _ := cmd.Flags().GetString("command")
			diffCmdSkipOnLabel, _ := cmd.Flags().GetBool("skip-on-label")
			lint, _ := cmd.Flags().GetBool("lint")

			if diffCmdFilename == "" {
				fmt.Println(cmd.Usage())
				return nil
			}

			return action.Diff(s, diffCmdFilename, diffCmdDiffer, diffCmdSkipOnLabel, lint)
		},
		Predictors: args.Predictors{
			"filename": args.PredictFile(),
			"command":  args.PredictFile(),
		},
	}

	c.cmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")
	c.cmd.Flags().StringP("command", "c", "diff", "External diff programm")
	c.cmd.Flags().BoolP("skip-on-label", "s", false, "Allow to skip resources based on their labels")
	c.cmd.Flags().BoolP("lint", "l", true, "Lint resources before diffing")

	return nil
}
