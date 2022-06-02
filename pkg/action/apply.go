package action

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

var (
	kindOrder = map[string]int{
		"Application": 0,
		"Deployment":  1,
	}
)

func Apply(s *service.Service, fileName string, skipOnLabel bool, lint bool) error {
	if lint {
		if err := Lint(s, fileName, true); err != nil {
			return err
		}
	}

	contents, err := api.GetFilesContentsToProcess(fileName)
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

	resourcesSorted := sortByKind(resources)

	for _, r := range resourcesSorted {
		if err := apply(r.ID, r.Res, skipOnLabel, s.Client); err != nil {
			return err
		}
	}

	return nil
}

type resAssoc struct {
	ID    string
	Res   api.Resource
	Order int
}

func sortByKind(resources map[string]api.Resource) []resAssoc {
	result := make([]resAssoc, len(resources))
	i := 0
	for resID, res := range resources {
		if res == nil {
			result[i] = resAssoc{
				resID,
				res,
				-1,
			}
			i++
			continue
		}
		kind := reflect.TypeOf(res).Elem().Name()
		var order int
		var exists bool
		if order, exists = kindOrder[kind]; kind == "" || !exists {
			order = -1
		}
		result[i] = resAssoc{
			resID,
			res,
			order,
		}
		i++
	}
	sort.SliceStable(result, func(i, j int) bool {
		s := result[i].Order
		t := result[j].Order

		// resources with no kind go last
		if s > -1 && t == -1 {
			return true
		}
		if t > -1 && s == -1 {
			return false
		}

		// fallback: sort by id
		if s == -1 && t == -1 || s == t {
			return result[i].ID < result[j].ID
		}

		// sort by kind
		return s < t
	})

	return result
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
		if err == nil {
			remoteLabels = app.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(app.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.Device:
		var device api.Device
		device, getErr = client.GetDeviceById(cl, resID, nil)
		if err == nil {
			remoteLabels = device.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(device.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.DeviceConfigMap:
		var dcm api.DeviceConfigMap
		dcm, getErr = getDeviceConfigMapByID(cl, resID)
		if err == nil {
			remoteLabels = dcm.Metadata.Labels
			v.Metadata.Labels = api.ComputeMetadataMap(dcm.Metadata.Labels, v.Metadata.Labels)
			update = v
		}
	case *api.Deployment:
		var deploy api.Deployment
		deploy, getErr = getDeploymentById(cl, resID, nil)
		if err == nil {
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
