package action

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func LabelDeployment(s *service.Service, ids []string) error {
	d, err := getDeploymentById(s.Client, ids[0], nil)
	if err != nil {
		return err
	}

	m, err := api.ParseMetadataMap(strings.Join(ids[1:], " "))
	if err != nil {
		return err
	}

	d.Metadata.Labels = api.ComputeMetadataMap(d.Metadata.Labels, m)

	message, err := updateResource(s.Client, d, d.Metadata.ID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
