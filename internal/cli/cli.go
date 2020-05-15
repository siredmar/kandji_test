package cli

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/service"
)

// CLI is the entry point to all commands
type CLI struct {
	root *clix.Command
	svc  *service.Service
}

// New returns a new CLI
func New(svc *service.Service, root cmd.CMD) (*CLI, error) {
	err := Init(svc, root)
	if err != nil {
		return nil, err
	}

	return &CLI{
		root: root.Command(),
		svc:  svc,
	}, nil
}

// Exec the CLI according to root command
func (c *CLI) Exec() error {
	return c.root.Execute()
}

// Init root CMD and all children
func Init(svc *service.Service, node cmd.CMD) error {
	if err := node.Init(svc); err != nil {
		return err
	}

	// wrap Command.Run to init services - not pretty but we can't do it earlier
	// i.e. in order to read the auth config we need to know the value of the
	// --config flag, which is not set until a command is run
	c := node.Command()
	r := c.Run
	if r != nil {
		c.Run = func(cmd *clix.Command, args []string) error {
			if err := svc.Init(); err != nil {
				return err
			}

			return r(cmd, args)
		}
	}

	for _, child := range node.Children() {
		if err := Init(svc, child); err != nil {
			return err
		}

		node.Command().AddCommand(child.Command())
	}

	// as clix has no global/persistent flags, we need to set them for each command
	node.Command().Flags().BoolVar(&svc.Config.UseStaging, "staging", false, "use staging env")
	node.Command().Flags().StringVar(&svc.Config.Profile, "profile", "", "profile to use")
	node.Command().Flags().StringVar(&svc.Config.ConfigFile, "config", "", "config file")

	return nil
}
