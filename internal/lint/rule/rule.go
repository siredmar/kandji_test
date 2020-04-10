package rule

import (
	"github.com/grid-x/gxctl/internal/lint/result"
)

type Rule interface {
	ID() string
	Desc() string
	Exec(interface{}) (*result.Result, error)
}
