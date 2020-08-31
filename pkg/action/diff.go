package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/ghodss/yaml"
	deviceApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	cleanupconfigApi "github.com/grid-x/ds-api-types/management/2019-12-10/cleanupconfigs"
	dockerconfigApi "github.com/grid-x/ds-api-types/management/2019-12-10/dockerconfigs"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

const (
	KNOWN_AFTER_APPLY = "(Known after apply)"
)

func Diff(s *service.Service, fileName string, diffCmd string) error {
	contents, err := api.GetFilesContentsToProcess(fileName)
	if err != nil {
		return err
	}

	for n, c := range contents {
		if err := diff(n, c, diffCmd, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func diff(filename string, content []byte, differ string, client *client.APIClient) error {
	res, resID, err := checkResourceFile(content, true)
	if err != nil {
		return err
	}

	var f1, f2 string
	if resID == KNOWN_AFTER_APPLY {
		// Looks like a new resource... Diff against empty file
		err, f1, f2 = writeDiffFiles(nil, res)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err != nil {
			return err
		}
	} else {
		// There is a resID - Check if res already exists
		var current interface{}
		var err error

		switch v := res.(type) {
		case api.Application:
			current, err = getApplicationById(client, v.Metadata.ID)
		case api.Device:
			var device api.Device
			device, err = getDeviceById(client, v.Metadata.ID, nil)
			device.Status = deviceApi.DeviceStatus{}
			current = device
		case api.Deployment:
			var deploy api.Deployment
			deploy, err = getDeploymentById(client, v.Metadata.ID, nil)
			deploy.Status = deploymentsApi.DeviceDeploymentStatus{}
			current = deploy
		case api.DockerConfig:
			var dc api.DockerConfig
			dc, err = getDockerConfigById(client, v.Metadata.ID, nil)
			dc.Status = dockerconfigApi.DockerConfigStatus{}
			current = dc
		case api.CleanupConfig:
			var cc api.CleanupConfig
			cc, err = getCleanupConfigById(client, v.Metadata.ID, nil)
			cc.Status = cleanupconfigApi.CleanupConfigStatus{}
			current = cc
		default:
			return errors.E(
				errors.NotImplemented,
				"Unsupported type",
			)
		}

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
		default: //couldnt run diff
			return err
		}
	}

	if len(output) != 0 {
		fmt.Printf("%s:\n", filename)
		fmt.Println(string(output))
	}
	return nil
}

func checkResourceFile(bytes []byte, readOnly bool) (interface{}, string, error) {
	resID, err := resolveIdentifierFromFile(bytes)
	if err != nil && readOnly {
		resID = KNOWN_AFTER_APPLY
	}

	application, err := api.NewApplication(bytes, true)
	if err == nil {
		application.Name = resID
		return application, resID, nil
	}

	deployment, err := api.NewDeployment(bytes, true)
	if err == nil {
		deployment.Metadata.ID = resID
		return deployment, resID, nil
	}

	device, err := api.NewDevice(bytes, true)
	if err == nil {
		device.Metadata.ID = resID
		return device, resID, nil
	}

	dockerConfig, err := api.NewDockerConfig(bytes, true)
	if err == nil {
		dockerConfig.Metadata.ID = resID
		return dockerConfig, resID, nil
	}

	cleanupConfig, err := api.NewCleanupConfig(bytes, true)
	if err == nil {
		cleanupConfig.Metadata.ID = resID
		return cleanupConfig, resID, nil
	}

	maintenanceTask, err := api.NewMaintenanceTask(bytes, true)
	if err == nil {
		maintenanceTask.Metadata.ID = resID
		return maintenanceTask, resID, nil
	}

	//Nothing found
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
