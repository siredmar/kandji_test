package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestResourceUniqueID(t *testing.T) {
	r := ResourceUniqueID{}

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
			res: &api.Application{
				Name: "foo",
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "app does not exist",
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
			res: &api.Application{
				Name: "goo",
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "app has no name",
			ctx:  nilCtx,
			res: &api.Application{
				Name: "",
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "app exists currently",
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
			res: &api.Application{
				Name: "foo",
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "app will exist",
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
			res: &api.Application{
				Name: "foo",
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "no deployments",
			ctx:  nilCtx,
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment does not exist",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
							Spec: deploymentsApi.DeviceDeploymentSpec{
								App: "foobar",
							},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "goo",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment does not have an ID",
			ctx:  nilCtx,
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment exists currently with the same spec",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
							Spec: deploymentsApi.DeviceDeploymentSpec{
								App: "foobar",
							},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
				Spec: deploymentsApi.DeviceDeploymentSpec{
					App: "foobar",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment exists currently with a different spec",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
							Spec: deploymentsApi.DeviceDeploymentSpec{
								App: "foobar",
							},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
				Spec: deploymentsApi.DeviceDeploymentSpec{
					App: "goobaz",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment will exist with the same spec",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
							Spec: deploymentsApi.DeviceDeploymentSpec{
								App: "foobar",
							},
						},
					},
				},
			},
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
				Spec: deploymentsApi.DeviceDeploymentSpec{
					App: "foobar",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "deployment will exist with a different spec",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
							Spec: deploymentsApi.DeviceDeploymentSpec{
								App: "foobar",
							},
						},
					},
				},
			},
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
				Spec: deploymentsApi.DeviceDeploymentSpec{
					App: "goobaz",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "no dcms",
			ctx:  nilCtx,
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "dcm does not exist",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
						},
					},
				},
			},
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "goo",
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "dcm does exist",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
						},
					},
				},
			},
			res: &api.DeviceConfigMap{
				Metadata: types.Metadata{
					ID: "foo",
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
