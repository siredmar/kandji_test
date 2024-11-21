package api

import (
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
		resources []ResAssoc
		want      []ResAssoc
	}{
		{
			desc: "Application before Deployment",
			resources: []ResAssoc{
				{ID: "foo", Res: &Application{}},
				{ID: "bar", Res: &Deployment{}},
				{ID: "config", Res: &DeviceConfigMap{}},
			},
			want: []ResAssoc{
				{
					ID:    "foo",
					Res:   &Application{},
					Order: 0,
				},
				{
					ID:    "config",
					Res:   &DeviceConfigMap{},
					Order: 1,
				},
				{
					ID:    "bar",
					Res:   &Deployment{},
					Order: 2,
				},
			},
		},
		{
			desc: "Applications (sorted by by Name) before Deployments (sorted by ID)",
			resources: []ResAssoc{
				{
					ID: "d",
					Res: &Application{
						Name: "d",
					},
				},
				{
					Res: &Application{
						Name: "c",
					},
					ID: "c",
				},
				{
					Res: &DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "f",
						},
					},
					ID: "f",
				},
				{
					Res: &DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "e",
						},
					},
					ID: "e",
				},
				{
					Res: &Deployment{
						Metadata: types.Metadata{
							ID: "b",
						},
					},
					ID: "b",
				},
				{
					Res: &Deployment{
						Metadata: types.Metadata{
							ID: "a",
						},
					},
					ID: "a",
				},
			},
			want: []ResAssoc{
				{
					ID: "c",
					Res: &Application{
						Name: "c",
					},
					Order: 0,
				},
				{
					ID: "d",
					Res: &Application{
						Name: "d",
					},
					Order: 0,
				},
				{
					ID: "e",
					Res: &DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "e",
						},
					},
					Order: 1,
				},
				{
					ID: "f",
					Res: &DeviceConfigMap{
						Metadata: types.Metadata{
							ID: "f",
						},
					},
					Order: 1,
				},
				{
					ID: "a",
					Res: &Deployment{
						Metadata: types.Metadata{
							ID: "a",
						},
					},
					Order: 2,
				},
				{
					ID: "b",
					Res: &Deployment{
						Metadata: types.Metadata{
							ID: "b",
						},
					},
					Order: 2,
				},
			},
		},
		{
			desc: "Deployments before unspecified",
			resources: []ResAssoc{
				{
					ID:  "b",
					Res: &Deployment{},
				},
				{
					ID:  "a",
					Res: &mockResource{},
				},
			},
			want: []ResAssoc{
				{
					ID:    "b",
					Res:   &Deployment{},
					Order: 2,
				},
				{
					ID:    "a",
					Res:   &mockResource{},
					Order: -1,
				},
			},
		},
		{
			desc:      "no resources",
			resources: []ResAssoc{},
			want:      []ResAssoc{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			sortByKind(tc.resources)
			if diff := cmp.Diff(tc.want, tc.resources); diff != "" {
				t.Errorf("sortByKind error (-want +got):\n%s", diff)
			}
		})
	}
}
