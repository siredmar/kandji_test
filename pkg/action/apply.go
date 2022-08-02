package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Apply(s *service.Service, fileName string, skipOnLabel bool, lint bool) error {
	if lint {
		if err := Lint(s, fileName, true); err != nil {
			return err
		}
	}

	resources, err := api.GetResources(fileName, false, false)
	if err != nil {
		return err
	}

	for _, r := range resources {
		if err := apply(r.ID, r.Res, skipOnLabel, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func apply(resID string, res api.Resource, skipOnLabel bool, cl *client.APIClient) error {
	token, err := cl.GetToken()
	if err != nil {
		return err
	}
	if token.IsCI() {
		err := withManagedMeta(res)
		if err != nil {
			return errors.E(
				errors.Internal,
				"Inject managed annotation",
				err,
			)
		}
	}

	if resID == "" {
		// Create
		return create(resID, res, cl)
	}

	// Update or Create
	var update api.Resource
	var remoteLabels map[string]string
	var getErr error

	switch v := res.(type) {
	case *api.Application:
		var app api.Application
		app, getErr = getApplicationById(cl, resID)
		if getErr == nil {
			remoteLabels = app.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(app.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.Device:
		var device api.Device
		device, getErr = client.GetDeviceById(cl, resID, nil)
		if getErr == nil {
			remoteLabels = device.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(device.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.DeviceConfigMap:
		var dcm api.DeviceConfigMap
		dcm, getErr = getDeviceConfigMapByID(cl, resID)
		if getErr == nil {
			remoteLabels = dcm.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(dcm.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.Deployment:
		var deploy api.Deployment
		deploy, getErr = getDeploymentById(cl, resID, nil)
		if getErr == nil {
			remoteLabels = deploy.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(deploy.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	default:
		return errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", res),
		)
	}

	// Check if api returned something else then 404, if so we need to exit
	if getErr != nil {
		e, ok := getErr.(*errors.Error)
		if !ok {
			return errors.E(errors.Other, getErr)
		}

		if e.Kind != errors.NotExists {
			return e
		}
	}

	if update == nil {
		// Create
		return create(resID, res, cl)
	}

	// Update
	// Skip on label
	if _, ok := remoteLabels[api.IgnoreLabel]; skipOnLabel && ok {
		fmt.Printf("%s:\n", resID)
		fmt.Println("Ignored due to label: " + api.IgnoreLabel)
		return nil
	}

	message, err := updateResource(cl, update, resID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
