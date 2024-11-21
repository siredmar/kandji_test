package printer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/PaesslerAG/jsonpath"

	api "github.com/grid-x/gxctl/pkg/api"
)

type PodsConsoleOutput struct {
	raw api.Pods
}
type PodsConsoleOutputWide struct {
	raw api.Pods
}

func (do PodsConsoleOutput) Inject(i api.Pods) PodsConsoleOutput {
	do.raw = i

	return do
}

func (do PodsConsoleOutputWide) Inject(i api.Pods) PodsConsoleOutputWide {
	do.raw = i

	return do
}

func (do PodsConsoleOutput) Map() []PodConsoleOutput {
	output := make([]PodConsoleOutput, len(do.raw.Pods))
	for _, e := range do.raw.Pods {
		output = append(output, PodConsoleOutput{}.Map(e))
	}

	return output
}

func (do PodsConsoleOutputWide) Map() []PodConsoleOutputWide {
	output := make([]PodConsoleOutputWide, len(do.raw.Pods))
	for _, e := range do.raw.Pods {
		output = append(output, PodConsoleOutputWide{}.Map(e))
	}

	return output
}

func (do PodsConsoleOutput) Sort(sortBy string) PodsConsoleOutput {
	out := sortPods(do.raw, sortBy)
	do.raw.Pods = out

	return do
}

func (do PodsConsoleOutputWide) Sort(sortBy string) PodsConsoleOutputWide {
	out := sortPods(do.raw, sortBy)
	do.raw.Pods = out

	return do
}

func sortPods(in api.Pods, sortBy string) []api.Pod {
	s := interface{}(nil)
	rawJSON, _ := json.Marshal(in)
	_ = json.Unmarshal(rawJSON, &s)

	if len(in.Pods) < 2 {
		return in.Pods
	}

	sortValues := make(map[string]interface{})

	for _, d := range in.Pods {
		sortValues[d.Metadata.ID], _ = jsonpath.Get(fmt.Sprintf("$..pods[?(@.metadata.id==\"%s\")].%s", d.Metadata.ID, sortBy), s)
	}

	var found bool
	sort.Slice(in.Pods, func(i, j int) bool {
		di := sortValues[in.Pods[i].Metadata.ID]
		dj := sortValues[in.Pods[j].Metadata.ID]

		if fmt.Sprintf("%v", di) != "[]" {
			found = true
		}
		return fmt.Sprintf("%v", di) < fmt.Sprintf("%v", dj)
	})

	if !found {
		s := fmt.Sprintf("Warning: Did not find jsonpath value: %s. Result is unsorted", sortBy)
		fmt.Println(s)
	}

	return in.Pods
}
