package deployment

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
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
		Use:     "deployment ID",
		Aliases: []string{"deployments", "deploy"},
		Short:   "update deployment",
		Run: func(cmd *clix.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument ID not found",
					[]string{"run 'gxctl update deployment --help' for usage"},
				)
			}

			if !api.IsDockerImageValid(args[0]) {
				return errors.E(
					errors.Invalid,
					"required argument IMAGE not valdi",
					[]string{"run 'gxctl update deployment --help' for usage"},
				)
			}

			updateDeploymentCmdImage, _ := cmd.Flags().GetString("image")
			updateDeploymentCmdApp, _ := cmd.Flags().GetString("app")
			updateDeploymentCmdSelector, _ := cmd.Flags().GetString("selector")

			if updateDeploymentCmdImage == "" && updateDeploymentCmdApp == "" && updateDeploymentCmdSelector == "" {
				return errors.E(
					errors.Invalid,
					"Nothing to update",
					[]string{"run 'gxctl update deployment --help' for usage"},
				)
			}
			return action.UpdateDeployment(
				s,
				args[0],
				updateDeploymentCmdImage,
				updateDeploymentCmdApp,
				updateDeploymentCmdSelector,
			)
		},
	}

	c.cmd.Flags().StringP("image", "i", "", "Image for the deployment")
	c.cmd.Flags().StringP("app", "a", "", "App for the deployment")
	c.cmd.Flags().StringP("selector", "s", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")

	return nil
}
