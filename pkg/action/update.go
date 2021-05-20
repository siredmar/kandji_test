package action

import (
	"fmt"

	types "github.com/grid-x/ds-api-types"
	devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Update(s *service.Service, updateCmdFilename string, lint bool) error {
	if lint {
		if err := Lint(s, updateCmdFilename, true); err != nil {
			return err
		}
	}

	contents, err := api.GetFilesContentsToProcess(updateCmdFilename)
	if err != nil {
		return err
	}

	resources := make(map[string]api.Resource, len(contents))
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

func update(resID string, res api.Resource, client *client.APIClient) error {
	var update api.Resource

	switch v := res.(type) {
	case *api.Application:
		app, err := getApplicationById(client, resID)
		if err != nil {
			return errors.E(
				errors.NotExists,
				"application not found",
			)
		}
		app.Metadata.Labels = api.ComputeMetadataMap(app.Metadata.Labels, v.Metadata.Labels)
		update = &app
	case *api.Device:
		device, err := getDeviceById(client, resID, nil)
		if err != nil {
			return errors.E(
				errors.NotExists,
				"device not found",
			)
		}
		device.Metadata.Labels = api.ComputeMetadataMap(device.Metadata.Labels, v.Metadata.Labels)
		update = &device
	case *api.DeviceConfigMap:
		dcm, err := getDeviceConfigMapByID(client, resID)
		if err != nil {
			return errors.E(
				errors.NotExists,
				"deviceconfigmap not found",
			)
		}
		dcm.Metadata.Labels = api.ComputeMetadataMap(dcm.Metadata.Labels, v.Metadata.Labels)
		update = &dcm

	case *api.Deployment:
		deploy, err := getDeploymentById(client, resID, nil)
		if err != nil {
			return errors.E(
				errors.NotExists,
				"deployment not found",
			)
		}
		deploy.Metadata.Labels = api.ComputeMetadataMap(deploy.Metadata.Labels, v.Metadata.Labels)
		update = &deploy
	default:
		return errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", res),
		)
	}

	message, err := updateResource(client, update, resID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func updateResource(client *client.APIClient, v api.Resource, id string, ids []string) (string, error) {
	resId := id
	if ids != nil {
		var err error
		resId, err = api.LookupID(id, ids)
		if err != nil {
			return "", err
		}
	}

	switch v := v.(type) {
	case *api.Application:
		return fmt.Sprintf("WARNING: Skipped application %s from update as applications cannot be updated", v.Metadata.ID), nil

	case *api.Device:
		in := api.UpdateDevice{}
		inSpec := devicesApi.UpdateSpec{}

		inSpec.MACAddress = v.Spec.MACAddress
		inSpec.MaintenanceWindow = v.Spec.MaintenanceWindow
		in.Spec = inSpec

		in.Metadata = &types.UpdateMetadata{}
		in.Metadata.Labels = v.Metadata.Labels

		response, err := client.PatchRequest(api.DevicesEndpoint, &in, resId)
		if err != nil {
			return "", err
		}

		device, err := api.NewDevice(response, false)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Device %s updated successfully", device.Metadata.ID), nil

	case *api.DeviceConfigMap:
		in := api.UpdateDeviceConfigMap{}
		inSpec := &v.Spec

		inSpec.Immutable = v.Spec.Immutable
		inSpec.Data = v.Spec.Data
		inSpec.BinaryData = v.Spec.BinaryData
		in.Spec = inSpec

		in.Metadata = types.UpdateMetadata{}
		in.Metadata.Labels = v.Metadata.Labels

		response, err := client.PatchRequest(api.DeviceConfigMapsEndpoint, &in, resId)
		if err != nil {
			return "", err
		}

		dcm, err := api.NewDeviceConfigMap(response, false)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("DeviceConfigMap %s updated successfully", dcm.Metadata.ID), nil

	case *api.Deployment:
		in := api.UpdateDeployment{}
		in.Spec = &v.Spec

		in.Metadata = types.UpdateMetadata{}
		in.Metadata.Labels = v.Metadata.Labels

		response, err := client.PatchRequest(api.DeploymentsEndpoint, &in, resId)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response, false)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s updated successfully", deployment.Metadata.ID), nil

	default:
		return "", errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", v),
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
