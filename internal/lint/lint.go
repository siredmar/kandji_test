package lint

import (
	"fmt"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/internal/lint/rule"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
)

// Lint the given resource
func Lint(ctx *context.Context, client *client.APIClient, fileName string, resource interface{}) ([]result.Result, error) {
	var rules []rule.Rule
	var results []result.Result

	switch resource.(type) {
	case api.Deployment:
		rules = append(rules,
			&rule.DeploymentSelectorUniqueMatchByDeviceID{},
			&rule.DeploymentSelectorSingleMatching{},
			&rule.DeploymentAppExists{},
		)
	}

	if len(rules) == 0 {
		fmt.Println("No matching rules")
		return nil, nil
	}

	for _, rule := range rules {
		result, err := rule.Exec(ctx, resource)
		if err != nil || result == nil {
			fmt.Printf("error while applying rule %v: %v\n", rule.ID(), err)
			return nil, err
		}
		result.SourceID = fileName
		result.Rule = rule
		results = append(results, *result)
	}

	return results, nil
}
