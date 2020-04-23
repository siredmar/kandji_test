package rule

import (
	"fmt"
	"testing"

	deployments "github.com/grid-x/ds-api-types/management/2019-12-10/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentAppExists(t *testing.T) {
	r := DeploymentAppExists{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "no applications",
			ctx:  nilCtx,
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "exists currently",
			ctx: &context.Context{
				Current: state.State{
					Applications: []api.Application{
						{
							Name: "foo",
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "will exist",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Applications: []api.Application{
						{
							Name: "foo",
						},
					},
				},
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "does not exist but will exist",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Applications: []api.Application{
						{
							Name: "foo",
						},
					},
				},
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "does not exist and will not exist",
			ctx: &context.Context{
				Current: state.State{
					Applications: []api.Application{
						{
							Name: "foo",
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "goo",
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
