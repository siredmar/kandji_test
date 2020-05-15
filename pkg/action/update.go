package action

import (
	"fmt"

	types "github.com/grid-x/ds-api-types"
	devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	cleanupConfigApi "github.com/grid-x/ds-api-types/management/2019-12-10/cleanupconfigs"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2019-12-10/deployments"
	dockerConfigApi "github.com/grid-x/ds-api-types/management/2019-12-10/dockerconfigs"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Update(s *service.Service, updateCmdFilename string) error {
	contents, err := api.GetFilesContentsToProcess(updateCmdFilename)
	if err != nil {
		return err
	}

	resources := make(map[string]interface{}, len(contents))
	for _, c := range contents {
		res, resID, err := checkResourceFile(c, false)
		if err != nil {
			return err
		}
		resources[resID] = res
	}

	for resID, res := range resources {
		if err := update(resID, res, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func update(resID string, res interface{}, client *client.APIClient) error {
	message, err := updateResource(client, res, resID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func updateResource(client *client.APIClient, v interface{}, id string, ids []string) (string, error) {
	resId := id
	if ids != nil {
		var err error
		resId, err = api.LookupID(id, ids)
		if err != nil {
			return "", err
		}
	}

	switch v := v.(type) {
	case api.Application:
		return fmt.Sprintf("WARNING: Skipped application %s from update as applications cannot be updated", v.Metadata.ID), nil
	case api.Device:
		in := devicesApi.UpdateRequest{}
		inSpec := devicesApi.UpdateSpec{}

		inSpec.MACAddress = v.Spec.MACAddress
		inSpec.MaintenanceWindow = v.Spec.MaintenanceWindow
		in.Spec = inSpec

		in.Metadata = &types.UpdateMetadata{}
		in.Metadata.Annotations = v.Metadata.Annotations
		in.Metadata.Annotations = v.Metadata.Labels

		response, err := client.PatchRequest(api.DevicesEndpoint, v, resId)
		if err != nil {
			return "", err
		}

		device, err := api.NewDevice(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Device %s updated successfully", device.Metadata.ID), nil
	case api.Deployment:
		in := deploymentsApi.UpdateRequest{}
		in.Spec = &v.Spec

		in.Metadata = types.UpdateMetadata{}
		in.Metadata.Labels = v.Metadata.Labels
		in.Metadata.Annotations = v.Metadata.Annotations

		response, err := client.PatchRequest(api.DeploymentsEndpoint, v, resId)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s updated successfully", deployment.Metadata.ID), nil
	case api.DockerConfig:
		in := dockerConfigApi.UpdateRequest{}
		in.Spec = &v.Spec

		in.Metadata.Labels = v.Metadata.Labels
		in.Metadata.Annotations = v.Metadata.Annotations

		response, err := client.PatchRequest(api.DockerConfigsEndpoint, v, resId)
		if err != nil {
			return "", err
		}

		config, err := api.NewDockerConfig(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("DockerConfig %s updated successfully", config.Metadata.ID), nil
	case api.CleanupConfig:
		in := cleanupConfigApi.UpdateRequest{}
		in.Spec = &v.Spec

		in.Metadata.Labels = v.Metadata.Labels
		in.Metadata.Annotations = v.Metadata.Annotations

		response, err := client.PatchRequest(api.CleanupConfigsEndpoint, v, resId)
		if err != nil {
			return "", err
		}

		config, err := api.NewCleanupConfig(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("CleanupConfig %s updated successfully", config.Metadata.ID), nil
	default:
		return "", errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}
}

func resolveIdentifierFromFile(bytes []byte) (string, error) {
	fullMeta, err := api.NewFullObjectMeta(bytes)
	if err != nil {
		return "", err
	}

	if fullMeta.Meta.Name != "" {
		return fullMeta.Meta.Name, nil
	}
	if fullMeta.Meta.Id != "" {
		return fullMeta.Meta.Id, nil
	}

	return "", fmt.Errorf("not found")
}
