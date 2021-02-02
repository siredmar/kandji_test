package action

import (
	"fmt"

	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func UpdateDeployment(
	s *service.Service,
	id string,
	updateDeploymentCmdImage string,
	updateDeploymentCmdApp string,
	updateDeploymentCmdSelector string,
) error {
	deployment, err := getDeploymentById(s.Client, id, nil)
	if err != nil {
		return err
	}

	if updateDeploymentCmdSelector != "" {
		matchByLabels, err := api.ParseMetadataMap(updateDeploymentCmdSelector)
		if err != nil {
			return err
		}

		deployment.Spec.Selector = deploymentsApi.Selector{
			MatchByLabels: matchByLabels,
		}
	}

	if updateDeploymentCmdImage != "" {
		name, err := api.GetDockerImageName(updateDeploymentCmdImage)
		if err != nil {
			return err
		}

		deployment.Spec.Template.Spec.Containers[0].Name = name
		deployment.Spec.Template.Spec.Containers[0].Image = updateDeploymentCmdImage
	}

	if updateDeploymentCmdApp != "" {
		deployment.Spec.App = updateDeploymentCmdApp
	}

	message, err := updateResource(s.Client, deployment, deployment.Metadata.ID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
