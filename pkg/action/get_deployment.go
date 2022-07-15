package action

import (
	"errors"
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"

	print "github.com/grid-x/gxctl/pkg/printer"
)

func GetDeployment(s *service.Service, deviceID string, outputType string, serial string, app string, sortBy string, ids []string) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
	}

	if deviceID != "" {
		//Get deployments by Device ID
		deployments, err := getDeploymentsByDeviceId(s.Client, deviceID)
		if err != nil {
			return err
		}

		deployments = filterByApp(deployments, app)

		if err := s.Printer.Print(deployments, printerConfig); err != nil {
			return err
		}
	} else if len(ids) > 0 {
		for _, a := range ids {
			deployment, err := getDeploymentById(s.Client, a, nil)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(deployment, printerConfig); err != nil {
				return err
			}
		}
	} else if serial != "" {
		responseList, err := getDeviceBySN(s.Client, serial)
		if err != nil {
			return err
		}

		if len(responseList) != 1 {
			return errors.New(fmt.Sprintf("unexpected length of responseList: %v", responseList))
		}

		var deviceID string

		for _, devices := range responseList {
			ids := devices.Device.GetIds()
			if len(ids) != 1 {
				return errors.New(fmt.Sprintf("unexpected length of devices: %v", len(ids)))
			}
			deviceID = ids[0]
		}

		// Get deployments by Device ID
		deployments, err := getDeploymentsByDeviceId(s.Client, deviceID)
		if err != nil {
			return err
		}

		deployments = filterByApp(deployments, app)

		if err := s.Printer.Print(deployments, printerConfig); err != nil {
			return err
		}

	} else {
		deployments, err := getDeployments(s.Client)
		if err != nil {
			return err
		}

		deployments = filterByApp(deployments, app)

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

func getDeploymentsByDeviceId(client *client.APIClient, deviceID string) (api.Deployments, error) {
	pods, err := getPodByDeviceId(client, deviceID)
	if err != nil {
		return api.Deployments{}, err
	}

	var deploymentIDs []string

	for _, pod := range pods.Pods {
		if pod.Metadata.Annotations == nil {
			continue
		}
		id, ok := pod.Metadata.Annotations["gridx.ai/deployment"]
		if ok {
			deploymentIDs = append(deploymentIDs, id)
		}
	}

	var deployments []api.Deployment

	for _, id := range deploymentIDs {
		deployment, err := getDeploymentById(client, id, deploymentIDs)
		if err != nil {
			return api.Deployments{}, err
		}

		deployments = append(deployments, deployment)
	}

	return api.Deployments{Deployments: deployments}, nil
}

func filterByApp(depls api.Deployments, filter string) api.Deployments {
	if filter == "" {
		return depls
	}
	retDepls := api.Deployments{Deployments: make([]api.Deployment, 0, len(depls.Deployments))}

	for _, depl := range depls.Deployments {
		if strings.EqualFold(depl.Spec.App, filter) {
			retDepls.Deployments = append(retDepls.Deployments, depl)

		}
	}
	return retDepls
}
