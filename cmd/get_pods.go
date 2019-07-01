package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/template"
)

type GetPods struct {
	Command *cobra.Command
}

func NewGetPods(parent *cobra.Command, client *client.APIClient, printer *print.Printer) *GetPods {
	var getPodsCmd = &cobra.Command{
		Use:                   "pod [NAME] [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"pods", "po"},
		Short:                 "get pod",
		Long:                  `Prints a list of all pods you have access to`,
		Example:               "# Get all pods \n  gxctl get pods\n\n  # Get information about an pod with abbreviation a1n \n  gxctl get pod a1n",
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")
			getCmdShowAll, _ := cmd.Flags().GetBool("all")

			printerConfig := print.Printconfig{
				OutputFormat: getCmdOutputType,
				ShowAll:      getCmdShowAll,
			}

			getCmdDeviceID, err := cmd.Flags().GetString("device-id")
			if err != nil {
				return err
			}

			if getCmdDeviceID != "" {
				//Get pods by Device ID
				pods, err := getPodByDeviceId(client, getCmdDeviceID)
				if err != nil {
					return err
				}

				if err := printer.Print(pods, printerConfig); err != nil {
					return err
				}
			} else if len(args) > 0 {
				//Lookup all existing pods to validate ids and autocomplete them if necessary
				pods, err := getPods(client)
				if err != nil {
					return err
				}
				podIds := pods.GetIds()

				//Get pods
				for _, a := range args {
					pod, err := getPodById(client, a, podIds)
					if err != nil {
						return err
					}

					if err := printer.Print(pod, printerConfig); err != nil {
						return err
					}
				}
			} else {
				//List pods
				pods, err := getPods(client)
				if err != nil {
					return err
				}

				if err := printer.Print(pods, printerConfig); err != nil {
					return err
				}
			}
			return nil
		},
	}

	getPodsCmd.SetHelpTemplate(template.HelpTemplate())
	getPodsCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(getPodsCmd)
	getPodsCmd.Flags().StringP("device-id", "d", "", "specify device id")
	getPodsCmd.Flags().BoolP("all", "", false, "show also unstarted pods")

	return &GetPods{
		Command: getPodsCmd,
	}
}

func getPods(client *client.APIClient) (api.Pods, error) {
	response, err := client.GetRequest(api.PodsEndpoint)
	if err != nil {
		return api.Pods{}, err
	}

	podList, err := api.NewPods(response)
	if err != nil {
		return podList, err
	}

	if podList.IsEmpty() {
		return podList, errors.ListNotFoundError("pods")
	}

	return podList, nil
}

func getPodById(client *client.APIClient, id string, podIds []string) (api.Pod, error) {
	podID, err := api.LookupID(id, podIds)
	if err != nil {
		return api.Pod{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.PodsEndpoint, podID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Pod{}, err
	}

	pod, err := api.NewPod(response)
	if err != nil {
		return pod, err
	}

	return pod, nil
}

func getPodByDeviceId(client *client.APIClient, id string) (api.Pods, error) {
	return api.Pods{}, errors.NotImplementedError("Getting Pods by Device-ID")
	/*
		endpoint := fmt.Sprintf("%s/%s/pods"", api.PodsEndpoint, id)
		response, err := apiClient.Request(endpoint)
		if err != nil {
			return api.Pods{}, err
		}
		podList, err := api.NewPods(response)

		if err != nil {
			return pods, err
		}

		if podList.IsEmpty() {
			return podList, errors.NotFoundError(errors.ErrorDetails{Command: "pods"})
		}

		return pods, nil
	*/
}
