package cmd

import (
	"github.com/spf13/cobra"

	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type GetMaintenance struct {
	Command *cobra.Command
}

func NewGetMaintenance(parent *cobra.Command) *GetMaintenance {
	var getMaintenanceCmd = &cobra.Command{
		Use:     "maintenance",
		Short:   "Get different maintenances",
		Long:    `TODO`,
		Example: "# Get all maintenance tasks \n  gxctl get maintenances\n\n  # Get information about an maintenance task with abbreviation a56 \n  gxctl get maintenance a56",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.NotImplementedError("Maintenance")
		},
	}

	getMaintenanceCmd.SetHelpTemplate(template.HelpTemplate())
	getMaintenanceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(getMaintenanceCmd)

	return &GetMaintenance{
		Command: getMaintenanceCmd,
	}
}
