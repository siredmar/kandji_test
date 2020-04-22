package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
)

type Rule interface {
	ID() string
	Desc() string
	Exec(ctx *context.Context, resource interface{}) (*result.Result, error)
}
