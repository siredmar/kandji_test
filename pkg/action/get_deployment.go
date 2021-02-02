package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"

	print "github.com/grid-x/gxctl/pkg/printer"
)

func GetDeployment(s *service.Service, outputType string, sortBy string, showDevices bool, ids []string) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
	}

	if len(ids) > 0 {
		for _, a := range ids {
			if showDevices {
				// Just show devices
				devices, err := getDeploymentDevices(s.Client, a, nil)
				if err != nil {
					return err
				}

				if err := s.Printer.Print(devices, printerConfig); err != nil {
					return err
				}
			} else {
				deployment, err := getDeploymentById(s.Client, a, nil)
				if err != nil {
					return err
				}

				if err := s.Printer.Print(deployment, printerConfig); err != nil {
					return err
				}
			}
		}
	} else {
		deployments, err := getDeployments(s.Client)
		if err != nil {
			return err
		}

		if outputType == "json" || outputType == "yaml" {
			if err := s.Printer.Print(api.Deployments{Deployments: deployments.Deployments}, printerConfig); err != nil {
				return err
			}
		} else {
			var label []api.Deployment
			var id []api.Deployment

			for _, d := range deployments.Deployments {
				if d.Spec.Selector.MatchByDeviceID != nil {
					id = append(id, d)
					continue
				}
				label = append(label, d)
			}
			fmt.Print("Deployments by ID\n\n")
			if err := s.Printer.Print(api.Deployments{Deployments: id}, printerConfig); err != nil {
				return err
			}
			fmt.Print("Deployments by Label\n\n")
			if err := s.Printer.Print(api.Deployments{Deployments: label}, printerConfig); err != nil {
				return err
			}
		}
	}
	return nil

}

func getDeployments(client *client.APIClient) (api.Deployments, error) {
	response, err := client.GetRequest(api.DeploymentsEndpoint)
	if err != nil {
		return api.Deployments{}, err
	}

	deploymentList, err := api.NewDeployments(response, false)
	if err != nil {
		return deploymentList, err
	}

	if deploymentList.IsEmpty() {
		return deploymentList, errors.E(
			errors.NotExists,
			"no deployments found",
		)
	}

	return deploymentList, nil
}

func getDeploymentById(client *client.APIClient, id string, deploymentIds []string) (api.Deployment, error) {
	if len(id) != 36 && deploymentIds == nil {
		deployments, err := getDeployments(client)
		if err != nil {
			return api.Deployment{}, err
		}
		deploymentIds = deployments.GetIds()
	}

	deploymentID, err := api.LookupID(id, deploymentIds)
	if err != nil {
		return api.Deployment{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DeploymentsEndpoint, deploymentID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Deployment{}, err
	}

	deployment, err := api.NewDeployment(response, false)
	if err != nil {
		return deployment, err
	}

	return deployment, nil
}

func getDeploymentDevices(client *client.APIClient, id string, deploymentIds []string) (api.Devices, error) {
	if len(id) != 36 && deploymentIds == nil {
		deployments, err := getDeployments(client)
		if err != nil {
			return api.Devices{}, err
		}
		deploymentIds = deployments.GetIds()
	}

	deploymentID, err := api.LookupID(id, deploymentIds)
	if err != nil {
		return api.Devices{}, err
	}

	endpoint := fmt.Sprintf("%s/%s/devices", api.DeploymentsEndpoint, deploymentID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Devices{}, err
	}

	devices, err := api.NewDevices(response, false)
	if err != nil {
		return devices, err
	}

	return devices, nil
}
