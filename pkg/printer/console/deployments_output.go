package printer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/PaesslerAG/jsonpath"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeploymentsConsoleOutput struct {
	raw api.Deployments
}
type DeploymentsConsoleOutputWide struct {
	raw api.Deployments
}

func (do DeploymentsConsoleOutput) Inject(i api.Deployments) DeploymentsConsoleOutput {
	do.raw = i

	return do
}

func (do DeploymentsConsoleOutputWide) Inject(i api.Deployments) DeploymentsConsoleOutputWide {
	do.raw = i

	return do
}

func (do DeploymentsConsoleOutput) Filter(showAll bool) DeploymentsConsoleOutput {
	return do
}

func (do DeploymentsConsoleOutputWide) Filter(showAll bool) DeploymentsConsoleOutputWide {
	return do
}

func (do DeploymentsConsoleOutput) Map() []DeploymentConsoleOutput {
	var output []DeploymentConsoleOutput
	for _, e := range do.raw.Deployments {
		output = append(output, DeploymentConsoleOutput{}.Map(e))
	}

	return output
}

func (do DeploymentsConsoleOutputWide) Map() []DeploymentConsoleOutputWide {
	var output []DeploymentConsoleOutputWide
	for _, e := range do.raw.Deployments {
		output = append(output, DeploymentConsoleOutputWide{}.Map(e))
	}

	return output
}

func (do DeploymentsConsoleOutput) Sort(sortBy string) DeploymentsConsoleOutput {
	out := sortDeployments(do.raw, sortBy)
	do.raw.Deployments = out

	return do
}

func (do DeploymentsConsoleOutputWide) Sort(sortBy string) DeploymentsConsoleOutputWide {
	out := sortDeployments(do.raw, sortBy)
	do.raw.Deployments = out

	return do
}

func sortDeployments(in api.Deployments, sortBy string) []api.Deployment {
	s := interface{}(nil)
	rawJson, _ := json.Marshal(in)
	json.Unmarshal(rawJson, &s)

	if len(in.Deployments) < 2 {
		return in.Deployments
	}

	sortValues := make(map[string]interface{})

	for _, d := range in.Deployments {
		sortValues[d.Metadata.ID], _ = jsonpath.Get(fmt.Sprintf("$..deployments[?(@.metadata.id==\"%s\")].%s", d.Metadata.ID, sortBy), s)
	}

	var found bool
	sort.Slice(in.Deployments, func(i, j int) bool {
		di := sortValues[in.Deployments[i].Metadata.ID]
		dj := sortValues[in.Deployments[j].Metadata.ID]

		if fmt.Sprintf("%v", di) != "[]" {
			found = true
		}
		return fmt.Sprintf("%v", di) < fmt.Sprintf("%v", dj)
	})

	if !found {
		s := fmt.Sprintf("Warning: Did not find jsonpath value: %s. Result is unsorted", sortBy)
		fmt.Println(s)
	}

	return in.Deployments
}
