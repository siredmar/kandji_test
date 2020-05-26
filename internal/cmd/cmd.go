package cmd

import (
	clix "github.com/go-clix/cli"

	"github.com/grid-x/gxctl/pkg/service"
)

// CMD is a top-level command
type CMD interface {
	Init(*service.Service) error
	Command() *clix.Command
	Children() []CMD
}
