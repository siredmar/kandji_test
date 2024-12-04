package list

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/service"
)

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

func (c *CMD) Init(s *service.Service) error {
	c.cmd = &clix.Command{
		Use:   "list",
		Short: "get device logs settings for all devices",
		Run: func(cmd *clix.Command, args []string) error {
			output, _ := c.cmd.Flags().GetString("output")
			return action.ListLogs(s, output)
		},
	}

	c.cmd.Flags().StringP("output", "o", "wide", "Print result in a different format. Must be one of: json|wide|yaml")

	return nil
}
