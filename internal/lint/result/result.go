package result

import (
	"fmt"
)

// Result contains the result of running a rule against a resource
type Result struct {
	SourceID string
	Rule     rule
	Pass     bool
	Have     interface{}
	Want     interface{}
}

type rule interface {
	ID() string
	Desc() string
}

func (r Result) String() string {
	if r.Pass {
		return fmt.Sprintf("PASS %v: %v", r.SourceID, r.Rule.ID())
	}

	prefix := fmt.Sprintf("FAIL %v: %v - %v", r.SourceID, r.Rule.ID(), r.Rule.Desc())

	if r.Want != nil && r.Have != nil {
		return fmt.Sprintf("%v. Want: %v, have %v", prefix, r.Want, r.Have)
	}
	if r.Have != nil {
		return fmt.Sprintf("%v. Have: %v", prefix, r.Have)
	}
	return prefix
}
