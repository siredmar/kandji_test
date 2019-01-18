package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
)

type Create struct {
	Command *cobra.Command
}

func NewCreate(parent *cobra.Command) *Create {
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create different resources",
		Long:  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			createCmdFilename, _ := cmd.Flags().GetString("filename")

			client := client.NewAPIClient()

			if createCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			jsonFile, err := os.Open(createCmdFilename)
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			bytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			res, err := checkResourceFile(bytes)
			if err != nil {
				return err
			}

			message, err := createResource(client, res)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	createCmd.Flags().StringP("filename", "f", "", "Filename to file to use to create the resource")
	parent.AddCommand(createCmd)

	return &Create{
		Command: createCmd,
	}
}

func checkResourceFile(bytes []byte) (interface{}, error) {
	deploymentCreate, err := api.NewCreateDeployment(bytes)
	if err != nil {
		return nil, err
	}
	if !deploymentCreate.IsEmpty() && deploymentCreate.IsValid() {
		return deploymentCreate, nil
	}

	applicationCreate, err := api.NewCreateApplication(bytes)
	if err != nil {
		return nil, err
	}
	if !applicationCreate.IsEmpty() && applicationCreate.IsValid() {
		return applicationCreate, nil
	}

	//Nothing found
	return nil, errors.InvalidFormat()
}

func createResource(client *client.APIClient, v interface{}) (string, error) {
	switch v := v.(type) {
	case api.CreateDeployment:
		response, err := client.PostRequest(api.DeploymentsEndpoint, v)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s created successfully", deployment.UUID), nil
	case api.CreateApplication:
		response, err := client.PostRequest(api.ApplicationsEndpoint, v)
		if err != nil {
			return "", err
		}

		application, err := api.NewApplication(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Application %s created successfully", application.Name), nil
	default:
		s := fmt.Sprintf("Creating resource of type %s.", v)
		return "", errors.NotImplementedError(s)
	}
}
