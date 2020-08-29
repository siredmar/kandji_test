package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentSelectorUniqueMatchByDeviceID(t *testing.T) {
	r := DeploymentSelectorUniqueMatchByDeviceID{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "does not exist",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: &goo,
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "no deployments",
			ctx:  nilCtx,
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: &foo,
					},
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "does exist with a different app",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "gooApp",
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
			desc: "does exist with the same ID and with the same app",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo-1234",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "fooApp",
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
			desc: "does exist with a different ID and with the same app",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "goo-5678",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "fooApp",
					Selector: deployments.Selector{
						MatchByDeviceID: &foo,
					},
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "will exist with the same ID and a different app",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo-1234",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "gooApp",
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
			desc: "will exist with a different ID and a different app",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "goo-5678",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "gooApp",
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
			desc: "will exist with the same ID and the same app",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo-1234",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "fooApp",
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
			desc: "will exist with a different ID and the same app",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "goo-5678",
							},
							Spec: deployments.DeviceDeploymentSpec{
								App: "fooApp",
								Selector: deployments.Selector{
									MatchByDeviceID: &foo,
								},
							},
						},
					},
				},
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo-1234",
				},
				Spec: deployments.DeviceDeploymentSpec{
					App: "fooApp",
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
