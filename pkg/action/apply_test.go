package action

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	types "github.com/grid-x/ds-api-types"

	"github.com/grid-x/gxctl/pkg/api"
)

type mockResource struct {}
func (m *mockResource) Meta() *types.Metadata {
	return nil
}

func TestSortByKind(t *testing.T) {
	testcases := []struct {
		desc      string
		resources map[string]api.Resource
		want      []resAssoc
	}{
		{
			desc: "Application before Deployment",
			resources: map[string]api.Resource{
				"foo": &api.Application{},
				"bar": &api.Deployment{},
			},
			want: []resAssoc{
				{
					"foo",
					&api.Application{},
					0,
				},
				{
					"bar",
					&api.Deployment{},
					1,
				},
			},
		},
		{
			desc: "Applications (sorted by by Name) before Deployments (sorted by ID)",
			resources: map[string]api.Resource{
				"d": &api.Application{
					Name: "d",
				},
				"c": &api.Application{
					Name: "c",
				},
				"b": &api.Deployment{
					Metadata: types.Metadata{
						ID: "b",
					},
				},
				"a": &api.Deployment{
					Metadata: types.Metadata{
						ID: "a",
					},
				},
			},
			want: []resAssoc{
				{
					"c",
					&api.Application{
						Name: "c",
					},
					0,
				},
				{
					"d",
					&api.Application{
						Name: "d",
					},
					0,
				},
				{
					"a",
					&api.Deployment{
						Metadata: types.Metadata{
							ID: "a",
						},
					},
					1,
				},
				{
					"b",
					&api.Deployment{
						Metadata: types.Metadata{
							ID: "b",
						},
					},
					1,
				},
			},
		},
		{
			desc: "Deployments before unspecified",
			resources: map[string]api.Resource{
				"b": &api.Deployment{},
				"a": &mockResource{},
			},
			want: []resAssoc{
				{
					"b",
					&api.Deployment{},
					1,
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
			resources: map[string]api.Resource{},
			want:      []resAssoc{},
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
