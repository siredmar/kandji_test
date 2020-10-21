package main

import (
	"fmt"
	"os"

	"github.com/grid-x/gxctl/internal/cli"
	"github.com/grid-x/gxctl/internal/cmd/root"
	"github.com/grid-x/gxctl/pkg/service"
)

func main() {
	service, err := service.New()
	if err != nil {
		exitError(err)
	}

	CLI, err := cli.New(service, root.New())
	if err != nil {
		exitError(err)
	}

	if err := CLI.Exec(); err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
