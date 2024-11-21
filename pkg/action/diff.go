package action

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

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
	KnownAfterApply = "(Known after apply)"
)

func Diff(s *service.Service, fileName string, diffCmd string, skipOnLabel bool, lint bool) error {
	if lint {
		if err := Lint(s, fileName, true); err != nil {
			return err
		}
	}

	token, err := s.Client.GetToken()
	if err != nil {
		return err
	}

	isCI := token.IsCI()

	resources, err := api.GetResources(fileName, true, true)
	if err != nil {
		return err
	}

	for _, r := range resources {
		if err := diff(r.Filename, r.Res, r.ID, diffCmd, skipOnLabel, isCI, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func diff(filename string, res api.Resource, resID string, differ string, skipOnLabel bool, isCI bool, client *client.APIClient) error {
	var f1, f2 string
	if resID == KnownAfterApply {
		// Looks like a new resource... Diff against empty file
		var err1, err2 error
		f1, err1 = writeObjectToDiffFile(nil)
		f2, err2 = writeObjectToDiffFile(res)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err1 != nil {
			return fmt.Errorf("failed to write an empty object to a diff file: %v", err1)
		}
		if err2 != nil {
			return fmt.Errorf("failed to write an object to a diff file: %v", err2)
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

		var err1, err2 error
		f1, err1 = writeObjectToDiffFile(current)
		f2, err2 = writeObjectToDiffFile(res)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err1 != nil {
			return fmt.Errorf("failed to write a current version of the object to a diff file: %v", err1)
		}
		if err2 != nil {
			return fmt.Errorf("failed to write a new version of the object to a diff file: %v", err2)
		}
	}

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
		app, err = getApplicationByID(cl, v.Metadata.ID)
		res = &app

	case *api.Device:
		var device api.Device
		device, err = client.GetDeviceByID(cl, v.Metadata.ID, nil)
		device.Status = deviceApi.DeviceStatus{}
		res = &device

	case *api.Deployment:
		var deploy api.Deployment
		deploy, err = getDeploymentByID(cl, v.Metadata.ID, nil)
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

func writeObjectToDiffFile(obj any) (string, error) {
	tmpFile, err := os.CreateTemp("/tmp", "rev")
	if err != nil {
		return "", fmt.Errorf("failed to create a temp file for an object: %v", err)
	}

	// if the object is nil, the empty file is used for comparison
	if obj == nil {
		return tmpFile.Name(), nil
	}

	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal an object to yaml: %v", err)
	}

	var yamlToMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &yamlToMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal a yaml into a map: %v", err)
	}

	flatMap := make(map[string]any)
	// to be able to see if the difference belongs to a nested structure or not
	flatten("", yamlToMap, ".", flatMap)
	// so the file always looks the same from the diff perspective
	sortedList := yamlToSortedList(flatMap)

	var sb strings.Builder
	for _, kv := range sortedList {
		fmt.Fprintf(&sb, "%s: %v\n", kv.key, kv.val)
	}

	if _, err := tmpFile.Write([]byte(sb.String())); err != nil {
		return "", err
	}

	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func flatten(prefix string, nested any, delimiter string, flatMap map[string]any) {
	switch value := nested.(type) {
	case map[string]any:
		for key, v := range value {
			newKey := key
			if prefix != "" {
				newKey = prefix + delimiter + key
			}
			flatten(newKey, v, delimiter, flatMap)
		}
	case []any:
		for i, v := range value {
			newKey := prefix + "[" + strconv.Itoa(i) + "]"
			flatten(newKey, v, delimiter, flatMap)
		}
	default:
		flatMap[prefix] = value
	}
}

type yamlAsKV struct {
	key string
	val any
}

func yamlToSortedList(yamlAsFlatMap map[string]any) []yamlAsKV {
	sortedList := make([]yamlAsKV, len(yamlAsFlatMap))
	i := 0
	for k, v := range yamlAsFlatMap {
		sortedList[i] = yamlAsKV{key: k, val: v}
		i++
	}

	slices.SortStableFunc(sortedList, func(a, b yamlAsKV) int { return cmp.Compare(a.key, b.key) })
	return sortedList
}
