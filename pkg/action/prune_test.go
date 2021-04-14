package action

import (
	"testing"

	types "github.com/grid-x/ds-api-types"
	"github.com/grid-x/gxctl/pkg/api"
)

func Test_shouldPrune(t *testing.T) {
	testcases := []struct {
		desc    string
		record  *record
		want    bool
		wantErr bool
	}{
		{
			desc:    "nil record",
			record:  nil,
			want:    false,
			wantErr: true,
		},
		{
			desc: "empty record",
			record: &record{
				fileName: "",
				local:    api.Deployment{},
				remote:   api.Deployment{},
			},
			want:    false,
			wantErr: false,
		},
		{
			desc: "local set, remote empty",
			record: &record{
				fileName: "deploy-foo",
				local: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-foo",
						Labels: map[string]string{
							"label-foo": "label-value-bar",
						},
						Annotations: map[string]string{
							"annotation-foo": "annotation-value-bar",
						},
					},
				},
				remote: api.Deployment{},
			},
			want:    true,
			wantErr: false,
		},
		{
			desc: "local empty, remote set",
			record: &record{
				fileName: "",
				local:    api.Deployment{},
				remote: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-foo",
						Labels: map[string]string{
							"label-foo": "label-value-bar",
						},
						Annotations: map[string]string{
							"annotation-foo": "annotation-value-bar",
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			desc: "local == remote",
			record: &record{
				fileName: "deploy-foo",
				local: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-foo",
						Labels: map[string]string{
							"label-foo": "label-value-bar",
						},
						Annotations: map[string]string{
							"annotation-foo": "annotation-value-bar",
						},
					},
				},
				remote: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-foo",
						Labels: map[string]string{
							"label-foo": "label-value-bar",
						},
						Annotations: map[string]string{
							"annotation-foo": "annotation-value-bar",
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			desc: "local != remote",
			record: &record{
				fileName: "deploy-foo",
				local: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-foo",
						Labels: map[string]string{
							"label-foo": "label-value-bar",
						},
						Annotations: map[string]string{
							"annotation-foo": "annotation-value-bar",
						},
					},
				},
				remote: api.Deployment{
					Metadata: types.Metadata{
						ID: "deploy-goo",
						Labels: map[string]string{
							"label-goo": "label-value-baz",
						},
						Annotations: map[string]string{
							"annotation-goo": "annotation-value-baz",
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			got, gotErr := shouldPrune("diff", false, true, "foo", tc.record)
			if tc.wantErr != (gotErr != nil) {
				t.Errorf("wantErr=%v, gotErr=%v", tc.wantErr, gotErr)
			}
			if tc.want != got {
				t.Errorf("want %v, got:\n%v", tc.want, got)
			}
		})
	}

}
