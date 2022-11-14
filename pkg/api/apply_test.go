package api

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	types "github.com/grid-x/ds-api-types"
)

type mockResource struct{}

func (m *mockResource) Meta() *types.Metadata {
	return nil
}

func TestSortByKind(t *testing.T) {
	testcases := []struct {
		desc      string
		resources map[string]Resource
		want      []ResAssoc
	}{
		{
			desc: "Application before Deployment",
			resources: map[string]Resource{
				"foo": &Application{},
				"bar": &Deployment{},
				"config": &DeviceConfigMap{},
			},
			want: []ResAssoc{
				{
					"foo",
					&Application{},
					0,
				},
				{
					"config",
					&DeviceConfigMap{},
					1,
				},
				{
					"bar",
					&Deployment{},
					2,
				},
			},
		},
		{
			desc: "Applications (sorted by by Name) before Deployments (sorted by ID)",
			resources: map[string]Resource{
				"d": &Application{
					Name: "d",
				},
				"c": &Application{
					Name: "c",
				},
				"f": &DeviceConfigMap{
					Metadata: types.Metadata{
						ID: "f",
					},
				},
				"e": &DeviceConfigMap{
					Metadata: types.Metadata{
						ID: "e",
					},
				},
				"b": &Deployment{
					Metadata: types.Metadata{
						ID: "b",
					},
				},
				"a": &Deployment{
					Metadata: types.Metadata{
						ID: "a",
					},
				},
			},
			want: []ResAssoc{
				{
					"c",
					&Application{
						Name: "c",
					},
					0,
				},
				{
					"d",
					&Application{
						Name: "d",
					},
					0,
				},
				{
					"e",
					&DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "e",
						},
					},
					1,
				},
				{
					"f",
					&DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "f",
						},
					},
					1,
				},
				{
					"a",
					&Deployment{
						Metadata: types.Metadata{
							ID: "a",
						},
					},
					2,
				},
				{
					"b",
					&Deployment{
						Metadata: types.Metadata{
							ID: "b",
						},
					},
					2,
				},
			},
		},
		{
			desc: "Deployments before unspecified",
			resources: map[string]Resource{
				"b": &Deployment{},
				"a": &mockResource{},
			},
			want: []ResAssoc{
				{
					"b",
					&Deployment{},
					2,
				},
				{
					"a",
					&mockResource{},
					-1,
				},
			},
		},
		{
			desc:      "no resources",
			resources: map[string]Resource{},
			want:      []ResAssoc{},
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			got := sortByKind(tc.resources)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("sortByKind error (-want +got):\n%s", diff)
			}
		})
	}
}