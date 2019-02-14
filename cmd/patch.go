package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type Patch struct {
	Command *cobra.Command
}

func NewPatch(parent *cobra.Command) *Patch {
	var patchCmd = &cobra.Command{
		Use:                   "patch [OPTIONS]",
		Short:                 "Patch different resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Patch a deployment resource from file \n  gxctl patch -f deployment_new.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			patchCmdFilename, _ := cmd.Flags().GetString("filename")

			client := client.NewAPIClient()

			if patchCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			jsonFile, err := os.Open(patchCmdFilename)
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			bytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			res, resId, err := checkPatchResourceFile(bytes)
			if err != nil {
				return err
			}

			message, err := patchResource(client, res, resId, nil)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	patchCmd.Flags().StringP("filename", "f", "", "Filename to file to use to patch the resource")

	patchCmd.SetHelpTemplate(template.HelpTemplate())
	patchCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(patchCmd)

	return &Patch{
		Command: patchCmd,
	}
}

func checkPatchResourceFile(bytes []byte) (interface{}, string, error) {
	resId := resolveIdentifierFromFile(bytes)

	deploymentPatch, err := api.NewPatchDeployment(bytes)
	if err != nil {
		return nil, "", err
	}
	if !deploymentPatch.IsEmpty() && deploymentPatch.IsValid() {
		return deploymentPatch, resId, nil
	}

	devicePatch, err := api.NewPatchDevice(bytes)
	if err != nil {
		return nil, "", err
	}
	if !devicePatch.IsEmpty() && devicePatch.IsValid() {
		return devicePatch, resId, nil
	}

	//Nothing found
	return nil, "", errors.InvalidFormat()
}

func patchResource(client *client.APIClient, v interface{}, id string, ids []string) (string, error) {
	resId := id
	if ids != nil {
		var err error
		resId, err = api.LookupID(id, ids)
		if err != nil {
			return "", err
		}
	}

	switch v := v.(type) {
	case api.PatchDevice:
		response, err := client.PatchRequest(api.DevicesEndpoint, v, resId)
		if err != nil {
			return "", err
		}

		device, err := api.NewDevice(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Device %s patched successfully", device.Metadata.ID), nil
	case api.PatchDeployment:
		response, err := client.PatchRequest(api.DeploymentsEndpoint, v, resId)
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

func resolveIdentifierFromFile(bytes []byte) string {
	fullMeta, err := api.NewFullObjectMeta(bytes)
	if err != nil {
		return ""
	}

	if fullMeta.Meta.Name != "" {
		return fullMeta.Meta.Name
	}
	if fullMeta.Meta.Id != "" {
		return fullMeta.Meta.Id
	}

	return ""
}
