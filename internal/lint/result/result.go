package result

import (
	"fmt"
)

// Result contains the result of running a rule against a resource
type Result struct {
	Rule rule
	Pass bool
}

type rule interface {
	ID() string
	Desc() string
}

func (r Result) String() string {
	if r.Pass {
		return fmt.Sprintf("%v: PASS", r.Rule.ID())
	}

	return fmt.Sprintf("%v: FAIL - %v", r.Rule.ID(), r.Rule.Desc())
}
