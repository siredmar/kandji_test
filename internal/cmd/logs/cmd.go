package logs

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/internal/cmd/logs/disable"
	"github.com/grid-x/gxctl/internal/cmd/logs/enable"
	"github.com/grid-x/gxctl/internal/cmd/logs/get"
	"github.com/grid-x/gxctl/internal/cmd/logs/list"
	"github.com/grid-x/gxctl/internal/cmd/logs/update"
	"github.com/grid-x/gxctl/pkg/service"
)

type CMD struct {
	cmd      *clix.Command
	children []cmd.CMD
}

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
		Use:   "logs",
		Short: "manage device logs settings",
		Args: args.Args{
			Validator: args.ValidateSingle("deviceID"),
		},
	}

	c.children = []cmd.CMD{
		disable.New(),
		enable.New(),
		get.New(),
		list.New(),
		update.New(),
	}

	return nil
}
