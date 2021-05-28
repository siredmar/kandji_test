package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"

	deviceConfigMapApi "github.com/grid-x/ds-api-types/management/2021-03-10/deviceconfigmaps"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

var False = func() *bool { x := false; return &x }()
var True = func() *bool { x := true; return &x }()

func TestDeviceConfigMapImmutable(t *testing.T) {
	r := DeviceConfigMapImmutable{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "dcm does not exist",
			ctx: nilCtx,
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "foo-dcm",
				},
				Spec: deviceConfigMapApi.DeviceConfigMapSpec{
					Data: map[string]string{
						"foo": "bar",
					},
				},
				Status: deviceConfigMapApi.DeviceConfigMapStatus{},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "dcm mutable",
			ctx: &context.Context{
				Current: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{ID: "foo-dcm"},
							Spec: deviceConfigMapApi.DeviceConfigMapSpec{
								Immutable: False,
								Data:       map[string]string{},
								BinaryData: map[string][]byte{},
							},
							Status: deviceConfigMapApi.DeviceConfigMapStatus{},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "foo-dcm",
				},
				Spec: deviceConfigMapApi.DeviceConfigMapSpec{
					Data: map[string]string{
						"foo": "bar",
					},
				},
				Status: deviceConfigMapApi.DeviceConfigMapStatus{},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "dcm immutable",
			ctx: &context.Context{
				Current: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{ID: "foo-dcm"},
							Spec: deviceConfigMapApi.DeviceConfigMapSpec{
								Immutable: True,
								Data:       map[string]string{},
								BinaryData: map[string][]byte{},
							},
							Status: deviceConfigMapApi.DeviceConfigMapStatus{},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "foo-dcm",
				},
				Spec: deviceConfigMapApi.DeviceConfigMapSpec{
					Immutable: True,
					Data: map[string]string{
						"foo": "bar",
					},
				},
				Status: deviceConfigMapApi.DeviceConfigMapStatus{},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			got, gotErr := r.Exec(tc.ctx, tc.res)

			didErr := false
			if tc.wantError && gotErr == nil || !tc.wantError && gotErr != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, gotErr)
				didErr = true
			}

			got.SourceID = "test"
			got.Rule = &r

			if tc.wantPass != got.Pass {
				t.Errorf("wanted pass=%v, got pass=%v", tc.wantPass, got.Pass)
				didErr = true
			}
			if tc.wantSkip != got.Skip {
				t.Errorf("wanted skip=%v, got skip=%v", tc.wantSkip, got.Skip)
				didErr = true
			}
			if didErr {
				t.Errorf("have: %v", got.Have)
			}
		})
	}
}
