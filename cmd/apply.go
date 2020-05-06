package cmd

import (
	"reflect"
	"sort"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

var (
	kindOrder = map[string]int{
		"Application": 0,
		"Deployment":  1,
	}
)

type Apply struct {
	Command *cobra.Command
}

func NewApply(parent *cobra.Command, client *client.APIClient) *Apply {
	var applyCmd = &cobra.Command{
		Use:                   "apply [OPTIONS]",
		Short:                 "apply resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Apply a resource from file \n  gxctl apply -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			applyCmdFilename, _ := cmd.Flags().GetString("filename")

			if applyCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(applyCmdFilename)
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

			resourcesSorted := sortByKind(resources)

			for _, r := range resourcesSorted {
				if err := apply(r.ID, r.Res, client); err != nil {
					return err
				}
			}

			return nil
		},
	}

	applyCmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")

	applyCmd.SetHelpTemplate(template.HelpTemplate())
	applyCmd.SetUsageTemplate(template.UsageTemplate())

	parent.AddCommand(applyCmd)

	return &Apply{
		Command: applyCmd,
	}
}

type resAssoc struct {
	ID    string
	Res   interface{}
	Order int
}

func sortByKind(resources map[string]interface{}) []resAssoc {
	result := make([]resAssoc, len(resources))
	i := 0
	for resID, res := range resources {
		kind := reflect.TypeOf(res).Name()
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

func apply(resID string, res interface{}, client *client.APIClient) error {
	if resID == "" {
		// Create
		return create(res, client)
	}

	// Update or Create
	found := false
	switch res.(type) {
	case api.Application:
		_, err := getApplicationById(client, resID)
		if err == nil {
			found = true
		}
	case api.Device:
		_, err := getDeviceById(client, resID, nil)
		if err == nil {
			found = true
		}
	case api.Deployment:
		_, err := getDeploymentById(client, resID, nil)
		if err == nil {
			found = true
		}
	case api.DockerConfig:
		_, err := getDockerConfigById(client, resID, nil)
		if err == nil {
			found = true
		}
	case api.CleanupConfig:
		_, err := getCleanupConfigById(client, resID, nil)
		if err == nil {
			found = true
		}
	default:
		return errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}

	if !found {
		// Create
		return create(res, client)
	}

	// Update
	return update(resID, res, client)
}
