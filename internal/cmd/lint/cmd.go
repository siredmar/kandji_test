package lint

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
		Use:   "lint",
		Short: "Lint resources by evaluating both desired (local files) and current (remote) state",
		Run: func(cmd *clix.Command, args []string) error {
			lintCmdFilename, _ := cmd.Flags().GetString("filename")

			if lintCmdFilename == "" {
				fmt.Println(cmd.Usage())
				return nil
			}

			return action.Lint(s, lintCmdFilename)
		},
		Predictors: args.Predictors{
			"filename": args.PredictFile(),
		},
	}

	c.cmd.Flags().StringP("filename", "f", "", "filename or directory of files to lint")

	return nil
}
