package cmd

import (
	"github.com/spf13/cobra"
)

type Root struct {
	Command *cobra.Command
}

func NewRoot() *Root {
	var rootCmd = &cobra.Command{
		Use:          "gxctl",
		SilenceUsage: true,
		Short:        "gxctl controls the gridX device services",
		Long:         `More info will come hier TODO`,
	}

	return &Root{
		Command: rootCmd,
	}
}
