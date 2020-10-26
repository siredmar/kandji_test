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

func (do PodsConsoleOutput) ShowAll(showAll bool) PodsConsoleOutput {
	out := filterPods(do.raw.Pods, showAll)
	do.raw.Pods = out

	return do
}

func (do PodsConsoleOutputWide) ShowAll(showAll bool) PodsConsoleOutputWide {
	out := filterPods(do.raw.Pods, showAll)
	do.raw.Pods = out

	return do
}

func (do PodsConsoleOutput) Map() []PodConsoleOutput {
	var output []PodConsoleOutput
	for _, e := range do.raw.Pods {
		output = append(output, PodConsoleOutput{}.Map(e))
	}

	return output
}

func (do PodsConsoleOutputWide) Map() []PodConsoleOutputWide {
	var output []PodConsoleOutputWide
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

func filterPods(in []api.Pod, showAll bool) []api.Pod {
	var r []api.Pod
	for _, e := range in {
		// Filter out pods which are not picked up yet if !showAll
		if !showAll {
			if e.Status.StartTime == nil {
				continue
			}
		}
		r = append(r, e)
	}

	return r
}

func sortPods(in api.Pods, sortBy string) []api.Pod {
	s := interface{}(nil)
	rawJson, _ := json.Marshal(in)
	json.Unmarshal(rawJson, &s)

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
