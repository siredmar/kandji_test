package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	printer "github.com/grid-x/gxctl/pkg/printer"
)

type GetPods struct {
	Command *cobra.Command
}

func NewGetPods(parent *cobra.Command) *GetPods {
	var getPodsCmd = &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod"},
		Short:   "get pods",
		Long:    `Prints a list of all pods you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdDeviceID, _ := cmd.Flags().GetString("device-id")
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			client := client.NewAPIClient()
			printer := printer.NewPrinter()

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

				err = printer.Print(pods, getCmdOutputType)
				if err != nil {
					return err
				}
			} else if len(args) > 0 {
				//Get pods
				for _, a := range args {
					pod, err := getPodById(client, a)
					if err != nil {
						return err
					}

					err = printer.Print(pod, getCmdOutputType)
					if err != nil {
						return err
					}
				}
			} else {
				//List pods
				pods, err := getPods(client)
				if err != nil {
					return err
				}

				err = printer.Print(pods, getCmdOutputType)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	parent.AddCommand(getPodsCmd)

	return &GetPods{
		Command: getPodsCmd,
	}
}

func getPods(client *client.APIClient) (api.Pods, error) {
	response, err := client.Request(api.PodsEndpoint)
	if err != nil {
		return api.Pods{}, err
	}

	podList, err := api.NewPods(response)
	if err != nil {
		return podList, err
	}

	if len(podList.Pods) == 0 {
		return podList, errors.NotFoundError(errors.ErrorDetails{Command: "pods"})
	}

	return podList, nil
}

func getPodById(client *client.APIClient, id string) (api.Pod, error) {
	endpoint := fmt.Sprintf("%s/%s", api.PodsEndpoint, id)
	response, err := client.Request(endpoint)
	if err != nil {
		return api.Pod{}, err
	}

	pod, err := api.NewPod(response)
	if err != nil {
		return pod, err
	}

	if pod.UUID != id {
		return pod, errors.NotFoundError(errors.ErrorDetails{Command: "pod", Id: id})
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
		pods, err := api.NewPods(response)

		if err != nil {
			return pods, err
		}

		if len(pods.Pods) == 0 {
			errors.NotFoundError(api.ErrorDetails{Command: "pods", Id: id})
			return pods, errors.New("Not Found")
		}

		return pods, nil
	*/
}
