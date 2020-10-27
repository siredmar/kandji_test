package setup

import (
	"fmt"

	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/internal/cmd"
	"github.com/grid-x/gxctl/pkg/service"
)

const sshConfig = `# Append the following to ~/.ssh/config:

Host *.gridbox-tunnel
  ProxyCommand gxctl sshnext tunnel --profile='*' $(echo %h | cut -d'.' -f1)
  ServerAliveInterval 30
  StrictHostKeyChecking no
  HashKnownHosts no
  RequestTTY Yes
  User root

Host *.gridbox
  User root
  HostName 127.0.0.1
  Port 22222
  UserKnownHostsFile /dev/null
  StrictHostKeyChecking no
  ProxyCommand ssh -o 'ForwardAgent yes' $(echo %n | cut -d'.' -f1).gridbox-tunnel 'ssh-add /keys/id_wssh_rsa && nc %h %p'
`

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
		Use:   "setup",
		Short: "Print instructions to setup OpenSSH aliases for easy tunneling to devices",
		Run: func(cmd *clix.Command, args []string) error {
			fmt.Println(sshConfig)
			return nil
		},
	}

	return nil
}
