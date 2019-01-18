package cmd

import (
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/spf13/cobra"
)

type GetMaintenance struct {
	Command *cobra.Command
}

func NewGetMaintenance(parent *cobra.Command) *GetMaintenance {
	var getMaintenanceCmd = &cobra.Command{
		Use:   "maintenance",
		Short: "Get different maintenances",
		Long:  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.NotImplementedError("Maintenance")
		},
	}

	parent.AddCommand(getMaintenanceCmd)

	return &GetMaintenance{
		Command: getMaintenanceCmd,
	}
}
