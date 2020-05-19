package delete

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/internal/cmd/delete/application"
	"github.com/grid-x/gxctl/internal/cmd/delete/deployment"
	"github.com/grid-x/gxctl/internal/cmd/delete/device"
	"github.com/grid-x/gxctl/internal/cmd/delete/maintenance"
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
		Use:   "delete",
		Short: "delete one or multiple resources from a file or directory",
		Run: func(cmd *clix.Command, args []string) error {
			deleteCmdFilename, _ := cmd.Flags().GetString("filename")

			if deleteCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			return action.Delete(s, deleteCmdFilename)
		},
		Predictors: args.Predictors{
			"filename": args.PredictFile(),
		},
	}

	c.children = []cmd.CMD{
		application.New(),
		deployment.New(),
		device.New(),
		maintenance.New(),
	}

	c.cmd.Flags().StringP("filename", "f", "", "Filename or directory of files")

	return nil
}
