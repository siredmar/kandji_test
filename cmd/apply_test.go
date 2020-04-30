package cmd

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	types "github.com/grid-x/ds-api-types"

	"github.com/grid-x/gxctl/pkg/api"
)

func TestSortByKind(t *testing.T) {
	testcases := []struct {
		desc      string
		resources map[string]interface{}
		want      []resAssoc
	}{
		{
			desc: "Application before Deployment",
			resources: map[string]interface{}{
				"foo": api.Application{},
				"bar": api.Deployment{},
			},
			want: []resAssoc{
				{
					"foo",
					api.Application{},
				},
				{
					"bar",
					api.Deployment{},
				},
			},
		},
		{
			desc: "Applications (sorted by by Name) before Deployments (sorted by ID)",
			resources: map[string]interface{}{
				"d": api.Application{
					Name: "d",
				},
				"c": api.Application{
					Name: "c",
				},
				"b": api.Deployment{
					Metadata: types.Metadata{
						ID: "b",
					},
				},
				"a": api.Deployment{
					Metadata: types.Metadata{
						ID: "a",
					},
				},
			},
			want: []resAssoc{
				{
					"c",
					api.Application{
						Name: "c",
					},
				},
				{
					"d",
					api.Application{
						Name: "d",
					},
				},
				{
					"a",
					api.Deployment{
						Metadata: types.Metadata{
							ID: "a",
						},
					},
				},
				{
					"b",
					api.Deployment{
						Metadata: types.Metadata{
							ID: "b",
						},
					},
				},
			},
		},
		{
			desc: "Deployments before unspecified",
			resources: map[string]interface{}{
				"b": api.Deployment{},
				"a": int(42),
			},
			want: []resAssoc{
				{
					"b",
					api.Deployment{},
				},
				{
					"a",
					42,
				},
			},
		},
		{
			desc:      "no resources",
			resources: map[string]interface{}{},
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
