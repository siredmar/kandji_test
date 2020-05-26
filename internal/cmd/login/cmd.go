package login

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
		Use:   "login",
		Short: "retrieve a login token to authenticate with the api",
		Run: func(cmd *clix.Command, args []string) error {
			openBrowser, _ := cmd.Flags().GetBool("auto")

			return action.Login(s, openBrowser)
		},
	}

	c.cmd.Flags().BoolP("auto", "a", true, "Do open the browser window automatically")

	return nil
}
