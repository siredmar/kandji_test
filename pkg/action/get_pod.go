package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetPod(s *service.Service, deviceID string, outputType string, sortBy string, ids []string) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
	}

	if deviceID != "" {
		//Get pods by Device ID
		pods, err := getPodByDeviceId(s.Client, deviceID)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(pods, printerConfig); err != nil {
			return err
		}
	} else if len(ids) > 0 {
		//Lookup all existing pods to validate ids and autocomplete them if necessary
		pods, err := getPods(s.Client)
		if err != nil {
			return err
		}
		podIds := pods.GetIds()

		//Get pods
		for _, a := range ids {
			pod, err := getPodById(s.Client, a, podIds)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(pod, printerConfig); err != nil {
				return err
			}
		}
	} else {
		//List pods
		pods, err := getPods(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(pods, printerConfig); err != nil {
			return err
		}
	}
	return nil
}

func getPods(client *client.APIClient) (api.Pods, error) {
	response, err := client.GetRequest(api.PodsEndpoint)
	if err != nil {
		return api.Pods{}, err
	}

	podList, err := api.NewPods(response, false)
	if err != nil {
		return podList, err
	}

	if podList.IsEmpty() {
		return podList, errors.E(
			errors.NotExists,
			"no pods found",
		)
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

	pod, err := api.NewPod(response, false)
	if err != nil {
		return pod, err
	}

	return pod, nil
}

func getPodByDeviceId(client *client.APIClient, id string) (api.Pods, error) {
	endpoint := fmt.Sprintf("%s/%s/pods", api.DevicesEndpoint, id)

	response, err := client.GetRequest(endpoint)
	if err != nil {
		if errors.IsKind(err, errors.NotExists) {
			return api.Pods{}, nil
		}

		return api.Pods{}, err
	}

	return api.NewPods(response, false)
}
