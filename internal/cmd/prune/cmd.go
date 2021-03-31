package prune

import (
	"fmt"

	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cli/args"
	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/action"
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
		Use:   "prune",
		Short: "prune resources",
		Run: func(cmd *clix.Command, args []string) error {
			fileName, _ := cmd.Flags().GetString("filename")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")
			diffCmd , _ := cmd.Flags().GetString("command")

			if fileName == "" {
				fmt.Println(cmd.Usage())
				return nil
			}

			if !dryRun {
				token, err := s.Client.GetToken()
				if err != nil {
					return err
				}
				if !token.IsCI() {
					return errors.E(err, "This command is meant to be used in CI exclusively")
				}
			}

			return action.Prune(s, dryRun, fileName, yes, diffCmd)
		},
		Predictors: args.Predictors{
			"filename": args.PredictFile(),
		},
	}

	c.cmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")
	c.cmd.Flags().StringP("command", "c", "diff", "External diff programm")
	c.cmd.Flags().Bool("dry-run", true, "Show diff only")
	c.cmd.Flags().Bool("yes", false, "Answer yes to all interactive prompts")
	c.cmd.Flags().MarkHidden("dry-run")
	c.cmd.Flags().MarkHidden("yes")

	return nil
}
