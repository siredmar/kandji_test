package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
)

type Patch struct {
	Command *cobra.Command
}

func NewPatch(parent *cobra.Command) *Patch {
	var patchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Patch different resources",
		Long:  `TODO`,
	}

	patchCmd.PersistentFlags().StringP("filename", "f", "", "Filename to file to use to create the resource")
	parent.AddCommand(patchCmd)

	return &Patch{
		Command: patchCmd,
	}
}

func checkPatchResourceFile(bytes []byte) (interface{}, error) {
	deploymentPatch, err := api.NewPatchDeployment(bytes)
	if err != nil {
		return nil, err
	}
	if !deploymentPatch.IsEmpty() && deploymentPatch.IsValid() {
		return deploymentPatch, nil
	}

	devicePatch, err := api.NewPatchDevice(bytes)
	if err != nil {
		return nil, err
	}
	if !devicePatch.IsEmpty() && devicePatch.IsValid() {
		return devicePatch, nil
	}

	//Nothing found
	return nil, errors.InvalidFormat()
}

func patchResource(client *client.APIClient, v interface{}, id string) (string, error) {
	switch v := v.(type) {
	case api.PatchDevice:
		response, err := client.PatchRequest(api.DevicesEndpoint, v, id)
		if err != nil {
			return "", err
		}

		device, err := api.NewDevice(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Device %s patched successfully", device.Metadata.ID), nil
	case api.PatchDeployment:
		response, err := client.PatchRequest(api.DeploymentsEndpoint, v, id)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s patched successfully", deployment.Metadata.ID), nil
	default:
		s := fmt.Sprintf("Creating resource of type %s.", v)
		return "", errors.NotImplementedError(s)
	}
}
