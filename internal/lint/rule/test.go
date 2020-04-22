package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

var (
	nilState = state.State{
		Applications: []api.Application{},
		Deployments:  []api.Deployment{},
	}
	nilCtx = &context.Context{
		Current: nilState,
		Desired: nilState,
	}
	foo = "foo"
	goo = "goo"
)
