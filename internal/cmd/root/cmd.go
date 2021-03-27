package root

import (
	clix "github.com/go-clix/cli"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/internal/cmd/apply"
	"github.com/grid-x/gxctl/internal/cmd/config"
	"github.com/grid-x/gxctl/internal/cmd/create"
	"github.com/grid-x/gxctl/internal/cmd/delete"
	"github.com/grid-x/gxctl/internal/cmd/diff"
	"github.com/grid-x/gxctl/internal/cmd/get"
	"github.com/grid-x/gxctl/internal/cmd/label"
	"github.com/grid-x/gxctl/internal/cmd/lint"
	"github.com/grid-x/gxctl/internal/cmd/login"
	"github.com/grid-x/gxctl/internal/cmd/prune"
	"github.com/grid-x/gxctl/internal/cmd/restart"
	"github.com/grid-x/gxctl/internal/cmd/ssh"
	"github.com/grid-x/gxctl/internal/cmd/update"
	"github.com/grid-x/gxctl/internal/cmd/validate"
	"github.com/grid-x/gxctl/internal/cmd/version"
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
		Use:   "gxctl",
		Short: "CLI for gridX Device Services (DS)",
	}

	c.children = []cmd.CMD{
		apply.New(),
		config.New(),
		create.New(),
		delete.New(),
		diff.New(),
		get.New(),
		label.New(),
		lint.New(),
		login.New(),
		prune.New(),
		restart.New(),
		ssh.New(),
		update.New(),
		validate.New(),
		version.New(),
	}

	// NOTE: global config flags are defined in internal/cli

	return nil
}
