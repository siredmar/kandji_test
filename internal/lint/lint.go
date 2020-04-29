package lint

import (
	"fmt"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/internal/lint/rule"
	"github.com/grid-x/gxctl/pkg/api"
)

// Lint the given resource
func Lint(ctx *context.Context, fileName string, resource interface{}) ([]result.Result, error) {
	var rules []rule.Rule
	var results []result.Result

	switch resource.(type) {
	case api.Application:
		rules = append(rules,
			&rule.ResourceUniqueID{},
		)
	case api.Deployment:
		rules = append(rules,
			&rule.DeploymentAppExists{},
			rule.NewDeploymentVolumeMountsAllowed([]string{"/proc"}),
			rule.NewDeploymentVolumesAllowed([]string{"/var/run/supervisor.sock"}),
			&rule.DeploymentSelectorSingleMatching{},
			&rule.DeploymentSelectorSpecified{},
			&rule.DeploymentSelectorUniqueMatchByDeviceID{},
			&rule.DeploymentSelectorDeviceExists{},
			&rule.DeploymentVolumesExist{},
			&rule.ResourceUniqueID{},
		)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("No matching rules")
	}

	for _, rule := range rules {
		result, err := rule.Exec(ctx, resource)
		if err != nil || result == nil {
			return nil, fmt.Errorf("error while applying rule %v: %v", rule.ID(), err)
		}
		result.SourceID = fileName
		result.Rule = rule
		results = append(results, *result)
	}

	return results, nil
}
