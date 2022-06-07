package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/ghodss/yaml"
	deviceApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"
	dcmApi "github.com/grid-x/ds-api-types/management/2021-03-10/deviceconfigmaps"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

const (
	KNOWN_AFTER_APPLY = "(Known after apply)"
)

func Diff(s *service.Service, fileName string, diffCmd string, skipOnLabel bool, lint bool) error {
	if lint {
		if err := Lint(s, fileName, true); err != nil {
			return err
		}
	}

	var isCI bool
	token, err := s.Client.GetToken()
	if err != nil {
		return err
	}
	if token.IsCI() {
		isCI = true
	}

	contents, err := api.GetFilesContentsToProcess(fileName)
	if err != nil {
		return err
	}

	for n, c := range contents {
		if !hasSupportedExtension(n) && fileName != n {
			continue
		}
		res, resID, err := checkResourceFile(c, true)
		if err != nil {
			return err
		}
		if err := diff(n, res, resID, diffCmd, skipOnLabel, isCI, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func diff(filename string, res api.Resource, resID string, differ string, skipOnLabel bool, isCI bool, client *client.APIClient) error {
	var f1, f2 string
	if resID == KNOWN_AFTER_APPLY {
		// Looks like a new resource... Diff against empty file
		err, f1, f2 := writeDiffFiles(nil, res)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err != nil {
			return err
		}
	} else {
		// There is a resID - Check if res already exists
		current, err := getResource(client, res)
		// Check if api returned something else then 404
		if err != nil {
			e, ok := err.(*errors.Error)
			if !ok {
				return errors.E(errors.Other, err)
			}

			if e.Kind != errors.NotExists {
				return e
			}

			// It's 404, set current to nil for diff
			current = nil
		}

		if current != nil && current.Meta() != nil {
			remoteLabels := current.Meta().Labels
			// Skip on label
			if _, ok := remoteLabels[api.IgnoreLabel]; skipOnLabel && ok {
				fmt.Printf("%s:\n", filename)
				fmt.Println("Ignored due to label: " + api.IgnoreLabel)
				return nil
			}
		}

		if isCI {
			err = withoutManagedMeta(current)
			if err != nil {
				return errors.E(errors.Internal, "remove managed meta", err)
			}
		}

		// Drop annotations if existing
		withoutAnnotations(current)
		withoutAnnotations(res)

		err, f1, f2 = writeDiffFiles(current, res)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err != nil {
			return err
		}
	}

	// Diff it
	output, err := exec.Command(differ, f1, f2).Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// this is just an exit code error, no worries
		default: // couldnt run diff
			return err
		}
	}

	if len(output) != 0 {
		fmt.Printf("%s:\n", filename)
		fmt.Println(string(output))
	}
	return nil
}

func getResource(cl *client.APIClient, req api.Resource) (api.Resource, error) {
	var res api.Resource
	var err error

	switch v := req.(type) {
	case *api.Application:
		var app api.Application
		app, err = getApplicationById(cl, v.Metadata.ID)
		res = &app

	case *api.Device:
		var device api.Device
		device, err = client.GetDeviceById(cl, v.Metadata.ID, nil)
		device.Status = deviceApi.DeviceStatus{}
		res = &device

	case *api.Deployment:
		var deploy api.Deployment
		deploy, err = getDeploymentById(cl, v.Metadata.ID, nil)
		deploy.Status = deploymentsApi.DeviceDeploymentStatus{}
		res = &deploy

	case *api.DeviceConfigMap:
		var dcm api.DeviceConfigMap
		dcm, err = getDeviceConfigMapByID(cl, v.Metadata.ID)
		dcm.Status = dcmApi.DeviceConfigMapStatus{}
		res = &dcm

	default:
		return nil, errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", res),
		)
	}
	return res, err
}

func checkResourceFile(bytes []byte, readOnly bool) (api.Resource, string, error) {
	resID, err := resolveIdentifierFromFile(bytes)
	if err != nil && readOnly {
		resID = KNOWN_AFTER_APPLY
	}

	application, err := api.NewApplication(bytes, true)
	if err == nil {
		application.Name = resID
		return &application, resID, nil
	}

	deployment, err := api.NewDeployment(bytes, true)
	if err == nil {
		deployment.Metadata.ID = resID
		return &deployment, resID, nil
	}

	device, err := api.NewDevice(bytes, true)
	if err == nil {
		device.Metadata.ID = resID
		return &device, resID, nil
	}

	dcm, err := api.NewDeviceConfigMap(bytes, true)
	if err == nil {
		dcm.Metadata.ID = resID
		return &dcm, resID, nil
	}

	maintenanceTask, err := api.NewMaintenanceTask(bytes, true)
	if err == nil {
		maintenanceTask.Metadata.ID = resID
		return &maintenanceTask, resID, nil
	}

	// Nothing found
	return nil, "", errors.E(errors.Invalid, "Unsupported type")
}

func writeDiffFiles(i1, i2 interface{}) (error, string, string) {
	rev1, err := ioutil.TempFile("/tmp", "rev2")
	if err != nil {
		return err, "", ""
	}

	rev2, err := ioutil.TempFile("/tmp", "rev2")
	if err != nil {
		return err, "", ""
	}

	if i1 != nil {

		b1, err := yaml.Marshal(i1)
		if err != nil {
			return err, "", ""
		}
		if _, err := rev1.Write(b1); err != nil {
			return err, "", ""
		}
		if err := rev1.Close(); err != nil {
			return err, "", ""
		}
	}

	if i2 != nil {
		b2, err := yaml.Marshal(i2)
		if err != nil {
			return err, "", ""
		}
		if _, err := rev2.Write(b2); err != nil {
			return err, "", ""
		}
		if err := rev2.Close(); err != nil {
			return err, "", ""
		}
	}

	return nil, rev1.Name(), rev2.Name()
}
