package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"
	// devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentSelectorDeviceExists(t *testing.T) {
	r := DeploymentSelectorDeviceExists{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "does exist",
			ctx: &context.Context{
				Current: state.State{
					Devices: []api.Device{
						{
							Metadata: types.Metadata{
								ID: foo,
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: &foo,
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "does not exist",
			ctx: &context.Context{
				Current: state.State{
					Devices: []api.Device{
						{
							Metadata: types.Metadata{
								ID: goo,
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: &foo,
					},
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			got, gotErr := r.Exec(tc.ctx, tc.res)
			got.SourceID = "test"
			got.Rule = &r

			didErr := false
			if tc.wantError && gotErr == nil || !tc.wantError && gotErr != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, gotErr)
				didErr = true
			}
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
