package ssh

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
		Use:   "ssh ID",
		Short: "ssh to different devices",
		Args: args.Args{
			args.ValidateSingle("ID"),
			args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			flagCommand, _ := cmd.Flags().GetString("command")
			silent, _ := cmd.Flags().GetBool("silent")

			return action.SSH(s, args[0], flagCommand, silent)
		},
		Predictors: args.Predictors{
			"command": args.PredictNil(),
		},
	}

	c.cmd.Flags().StringP("command", "c", "", "specify a command to issue")
	c.cmd.Flags().BoolP("silent", "s", false, "don't print any output during connection setup")

	return nil
}
