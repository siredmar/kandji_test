package action

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func LabelDevice(s *service.Service, ids []string) error {
	d, err := getDeviceById(s.Client, ids[0], nil)
	if err != nil {
		return err
	}

	m, err := api.ParseMetadataMap(strings.Join(ids[1:], " "))
	if err != nil {
		return err
	}

	for k, v := range m {
		if d.Metadata.Labels == nil {
			d.Metadata.Labels = make(map[string]string)
		}
		if strings.HasSuffix(k, "-") {
			// delete original value in map if label with deletion suffix was found
			delete(d.Metadata.Labels, k[:len(k)-1])
		}
		d.Metadata.Labels[k] = v
	}

	message, err := updateResource(s.Client, &d, d.Metadata.ID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
