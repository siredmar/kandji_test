package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"

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
			desc: "app does not exist",
			ctx:  nilCtx,
			res: api.Application{
				Name: "foo",
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "app has no name",
			ctx:  nilCtx,
			res: api.Application{
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
			res: api.Application{
				Name: "foo",
			},
			wantPass:  false,
			wantSkip:  false,
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
			res: api.Application{
				Name: "foo",
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "deployment does not exist",
			ctx:  nilCtx,
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "deployment does not have an ID",
			ctx:  nilCtx,
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "deployment exists currently",
			ctx: &context.Context{
				Current: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
						},
					},
				},
				Desired: nilState,
			},
			res: api.Deployment{
				Metadata: types.Metadata{
					ID: "foo",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "deployment will exist",
			ctx: &context.Context{
				Current: nilState,
				Desired: state.State{
					Deployments: []api.Deployment{
						{
							Metadata: types.Metadata{
								ID: "foo",
							},
						},
					},
				},
			},
			res: api.Deployment{
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

			if tc.wantError && gotErr == nil || !tc.wantError && gotErr != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, gotErr)
			}
			if tc.wantPass != got.Pass {
				t.Errorf("wanted pass=%v, got pass=%v", tc.wantPass, got.Pass)
			}
			if tc.wantSkip != got.Skip {
				t.Errorf("wanted skip=%v, got skip=%v", tc.wantSkip, got.Skip)
			}
		})
	}
}
