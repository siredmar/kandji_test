package api

import (
	"fmt"
	"testing"
)

type computeMetadataMapTestcase struct {
	current map[string]string
	update  map[string]string
	output  map[string]string
}

func TestComputeMetadataMap(t *testing.T) {
	testcases := []computeMetadataMapTestcase{
		{
			current: map[string]string{
				"key": "value",
			},
			update: map[string]string{
				"key": "value",
			},
			output: map[string]string{
				"key": "value",
			},
		},
		{
			current: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
			update: map[string]string{
				"key":  "foobar",
				"key2": "value2",
			},
			output: map[string]string{
				"key":  "foobar",
				"key2": "value2",
			},
		},
		{
			current: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
			update: map[string]string{
				"key": "value",
			},
			output: map[string]string{
				"key":   "value",
				"key2-": "",
			},
		},
		{
			current: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
			update: map[string]string{
				"key":   "value",
				"key2-": "",
			},
			output: map[string]string{
				"key":   "value",
				"key2-": "",
			},
		},
	}

	for _, tc := range testcases {
		res := ComputeMetadataMap(tc.current, tc.update)

		if fmt.Sprint(res) != fmt.Sprint(tc.output) {
			t.Errorf("Result does not match. Expected: %s, Got: %s", tc.output, res)
		}
	}
}
