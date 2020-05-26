package action

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func LabelDevice(s *service.Service, ids []string) error {
	devices, err := getDevices(s.Client)
	if err != nil {
		return err
	}
	deviceIDs := devices.GetIds()

	d, err := getDeviceById(s.Client, ids[0], deviceIDs)
	if err != nil {
		return err
	}

	m, err := api.ParseMetadataMap(strings.Join(ids[1:], " "))
	if err != nil {
		return err
	}

	d.Metadata.Labels = m

	message, err := updateResource(s.Client, d, d.Metadata.ID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
