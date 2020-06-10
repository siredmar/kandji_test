package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetApplication(s *service.Service, outputType string, ids []string) error {
	printerConfig := print.Printconfig{
		OutputFormat: outputType,
	}

	if len(ids) > 0 {
		//Get multiple application
		for _, id := range ids {
			application, err := getApplicationById(s.Client, id)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(application, printerConfig); err != nil {
				return err
			}
		}
	} else {
		//List all applications
		applications, err := getApplications(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(applications, printerConfig); err != nil {
			return err
		}
	}
	return nil
}

func getApplications(client *client.APIClient) (api.Applications, error) {
	response, err := client.GetRequest(api.ApplicationsEndpoint)
	if err != nil {
		return api.Applications{}, err
	}

	applicationList, err := api.NewApplications(response, false)
	if err != nil {
		return applicationList, err
	}

	if applicationList.IsEmpty() {
		return applicationList, errors.E(
			errors.NotExists,
			"no applications found",
		)
	}

	return applicationList, nil
}

func getApplicationById(client *client.APIClient, id string) (api.Application, error) {
	endpoint := fmt.Sprintf("%s/%s", api.ApplicationsEndpoint, id)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Application{}, err
	}

	application, err := api.NewApplication(response, false)
	if err != nil {
		return application, err
	}

	return application, nil
}
