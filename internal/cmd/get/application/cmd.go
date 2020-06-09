package application

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
		Use:     "application [NAME]",
		Aliases: []string{"app", "apps"},
		Short:   "get application(s)",
		Long:    "print a list of all applications you have access to",
		Args: args.Args{
			args.ValidateNil(),
			args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			outputType, _ := cmd.Flags().GetString("output")
			if err := action.GetApplication(s, outputType, args); err != nil {
				return err
			}

			return nil
		},
		Predictors: args.Predictors{
			"output": args.PredictOutputType(),
		},
	}

	c.cmd.Flags().StringP("output", "o", "", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
