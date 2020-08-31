package rule

import (
	"fmt"
	"testing"

	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentSelectorSpecified(t *testing.T) {
	r := DeploymentSelectorSpecified{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "only MatchByDeviceID",
			ctx:  nilCtx,
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
			desc: "only MatchByLabels",
			ctx:  nilCtx,
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByLabels: map[string]string{
							"goo": "baz",
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "both",
			ctx:  nilCtx,
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: &foo,
						MatchByLabels: map[string]string{
							"goo": "baz",
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "none",
			ctx:  nilCtx,
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{},
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
