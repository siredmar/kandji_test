package apply

import (
	clix "github.com/go-clix/cli"

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
		Use:   "apply",
		Short: "apply resources",
		Run: func(cmd *clix.Command, args []string) error {
			applyCmdFilename, _ := cmd.Flags().GetString("filename")
			if applyCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			return action.Apply(s, applyCmdFilename)
		},
	}

	c.cmd.Flags().StringP("filename", "f", "", "filename or directory to file to use to create the resource")

	return nil
}
