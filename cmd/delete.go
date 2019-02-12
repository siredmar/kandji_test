package cmd

import (
	"github.com/spf13/cobra"

	template "github.com/grid-x/gxctl/pkg/template"
)

type Delete struct {
	Command *cobra.Command
}

func NewDelete(parent *cobra.Command) *Delete {
	var deleteCmd = &cobra.Command{
		Use:   "delete [OPTIONS]",
		Short: "Delete a resources",
		Long:  `TODO`,
	}

	deleteCmd.SetHelpTemplate(template.HelpTemplate())
	deleteCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteCmd)

	return &Delete{
		Command: deleteCmd,
	}
}
