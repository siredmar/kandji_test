package tunnel

import (
	"fmt"
	"os"

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
		Use:   "tunnel SN",
		Short: "tunnel stdin to device sshd",
		Long: `Tunnel std file handles to device sshd, for usage with OpenSSH ProxyCommand directive.

For example, to connect to device with SN D123-FOO-…, use:
  gxctl ssh setup # run once and follow instructions
  ssh D123-FOO.gridbox
`,
		Args: args.Args{
			args.ValidateSingle("SN"),
			args.PredictNil(),
		},
		Run: func(cmd *clix.Command, args []string) error {
			info, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("cant check for stdin")
			}
			if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
				fmt.Printf("%s\n\n%s\n", cmd.Usage(), cmd.Long)
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")

			err = action.SSHTunnel(s, quiet, args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			return nil
		},
		Predictors: args.Predictors{
			"SN": args.PredictNil(),
		},
	}

	c.cmd.Flags().BoolP("quiet", "q", false, "do not report progress")

	return nil
}
