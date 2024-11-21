package filter

import (
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

type LabelFilter struct {
	query     string
	selectors [][]string
}

func NewLabelFilter(query string) *LabelFilter {
	subqueries := strings.Split(query, ",")
	selectors := make([][]string, len(subqueries))

	for i, q := range subqueries {
		keyValue := strings.Split(q, "=")
		if len(keyValue) < 2 {
			selectors[i] = []string{q}
		} else {
			selectors[i] = keyValue
		}
	}

	return &LabelFilter{
		query:     query,
		selectors: selectors,
	}
}

func (f *LabelFilter) Eval(in interface{}) (bool, error) {
	device, ok := in.(api.Device)

	if !ok {
		return false, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	if f.query == "" {
		return true, nil
	}

	labels := device.Metadata.Labels

	for _, s := range f.selectors {
		if len(s) < 2 {
			// check if key exists only
			if _, ok := labels[s[0]]; !ok {
				return false, nil
			}
		} else {
			// check if key has value
			if v := labels[s[0]]; v != s[1] {
				return false, nil
			}
		}
	}

	return true, nil

}
