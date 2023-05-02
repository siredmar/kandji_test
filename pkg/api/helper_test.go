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

func TestGetResources(t *testing.T) {
	testcases := []struct {
		loc                      string
		readOnly                 bool
		checkForExtensionSupport bool
		want                     []ResAssoc
		wantErr                  bool
	}{
		{
			loc:                      "testdata/test1.yaml",
			readOnly:                 false,
			checkForExtensionSupport: false,
			want: []ResAssoc{
				{
					ID: "691573be-a7d8-47bb-9aea-0337009a981d", Res: &DeviceConfigMap{}, Order: 1,
				},
				{
					ID: "35cc20b0-94ee-475f-9a0e-8b53a08f2143", Res: &Deployment{}, Order: 2,
				},
			},
			wantErr: false,
		},
		{
			loc:                      "testdata/device.json",
			readOnly:                 false,
			checkForExtensionSupport: true,
			want: []ResAssoc{
				{
					ID: "411cbc75-44ed-4c94-82cd-272f337cf371", Res: &Device{}, Order: -1,
				},
			},
		},
		{
			loc:                      "testdata/test2.yaml",
			readOnly:                 true,
			checkForExtensionSupport: true,
			wantErr:                  true,
		},
		{
			loc:                      "testdata/test3.yaml",
			readOnly:                 false,
			checkForExtensionSupport: true,
			want: []ResAssoc{
				{
					ID:    "ds-metric-agent",
					Res:   &Application{},
					Order: 0,
				},
			},
		},
	}

	for i, tc := range testcases {
		sortedRes, err := GetResources(tc.loc, tc.readOnly, tc.checkForExtensionSupport)
		if !tc.wantErr && err != nil {
			t.Fatalf("case %d: unexpected error: %+v", i, err)
		} else if tc.wantErr && err == nil {
			t.Fatalf("case %d: expected error but did not get one", i)
		}

		// Check if resources has been split if multiple resources exist in the file
		if len(sortedRes) != len(tc.want) {
			t.Errorf("case %d: Result does not match. Expected: %d, Got: %d", i, len(tc.want), len(sortedRes))
		}

		for n, res := range sortedRes {
			if res.ID != tc.want[n].ID {
				t.Errorf("case %d: Result does not match. Expected: %s, Got: %s", i, tc.want[n].ID, res.ID)
			}

			if res.Order != tc.want[n].Order {
				t.Errorf("case %d: Result does not match. Expected: %d, Got: %d", i, tc.want[n].Order, res.Order)
			}
		}
	}
}
