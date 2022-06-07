package rule

import (
	"errors"
	"fmt"
	"testing"

	// devicesApi "github.com/grid-x/ds-api-types/management/2019-06-13/device"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

type mockClient struct {
	err error
}

func (m mockClient) GetDeviceById(id string) (api.Device, error) {
	return api.Device{}, m.err
}

func TestDeploymentSelectorDeviceExists(t *testing.T) {
	r := DeploymentSelectorDeviceExists{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		deviceErr error
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "does exist",
			ctx: &context.Context{
				Current: state.State{},
				Desired: nilState,
			},
			deviceErr: nil, // no error -> assumed the device exists
			res: &api.Deployment{
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
				Current: state.State{},
				Desired: nilState,
			},
			deviceErr: errors.New("does not exist!"),
			res: &api.Deployment{
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
			tc.ctx.Cl = mockClient{tc.deviceErr}
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
