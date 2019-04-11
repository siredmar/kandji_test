package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type CreateMaintenance struct {
	Command *cobra.Command
}

func NewCreateMaintenance(parent *cobra.Command, client *client.APIClient) *CreateMaintenance {
	var createMaintenanceCmd = &cobra.Command{
		Use:                   "maintenance TYPE=restart/shutdown --selector=SELECTOR [OPTIONS]",
		Short:                 "Creates an maintenance task",
		Aliases:               []string{"maintenances"},
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Create an restart maintenance task for matchByLabel selector gridx.de/channel=stable \n  gxctl create maintenance -s gridx.de/channel=stable",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("TYPE", "gxctl create maintenance -h")
			}

			if strings.ToLower(args[0]) != "restart" && strings.ToLower(args[0]) != "shutdown" {
				//Todo use InvalidParameter
				return errors.MissingParameter("TYPE", "gxctl create maintenance -h")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			createMaintenanceCmdType := strings.ToLower(args[0])
			createMaintenanceCmdSelector, _ := cmd.Flags().GetString("selector")

			if createMaintenanceCmdSelector == "" {
				return errors.MissingParameter("SELECTOR", "gxctl create deployment -h")
			}

			var taskType mainv1beta1.MaintenanceTaskType
			if createMaintenanceCmdType == "restart" {
				taskType = mainv1beta1.MaintenanceTaskTypeRestart
			} else if createMaintenanceCmdType == "shutdown" {
				taskType = mainv1beta1.MaintenanceTaskTypeShutdown
			}

			matchByLabels := make(map[string]string)
			if createMaintenanceCmdSelector != "" {
				labels := strings.Split(createMaintenanceCmdSelector, " ")
				for _, pair := range labels {
					z := strings.Split(pair, "=")
					matchByLabels[z[0]] = z[1]
				}
			}

			d := api.CreateMaintenanceTask{Spec: &mainv1beta1.MaintenanceTaskSpec{
				Type: taskType,
				Selector: appsv1beta1.Selector{
					MatchByLabels: matchByLabels,
				},
			}}

			message, err := createResource(client, d)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	createMaintenanceCmd.Flags().StringP("selector", "s", "", "A space seperated list of labels to match a device eg. gridx.de/channel=stable gridx.de/area=west-1")

	createMaintenanceCmd.SetHelpTemplate(template.HelpTemplate())
	createMaintenanceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(createMaintenanceCmd)

	return &CreateMaintenance{
		Command: createMaintenanceCmd,
	}
}
