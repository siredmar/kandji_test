package args

import (
	"fmt"

	clix "github.com/go-clix/cli"
	"github.com/posener/complete"

	"github.com/grid-x/gxctl/pkg/errors"
)

type Predictor complete.Predictor
type Predictors map[string]complete.Predictor
type Args clix.Args

// PredictFile returns the default shell file predictor
func PredictFile() Predictor {
	return complete.PredictFiles("*")
}

// PredictOutputType predicts one of the output types
func PredictOutputType() Predictor {
	return complete.PredictSet("json", "yaml", "wide")
}

// PredictNil predicts nothing
func PredictNil() Predictor {
	return complete.PredictFunc(func(args complete.Args) []string {
		return nil
	})
}

// ValidateNil passes iff no arg is specified
func ValidateNil() clix.ValidateFunc {
	return func(args []string) error {
		if len(args) > 0 {
			return errors.E(
				errors.Invalid,
				"no arguments allowed",
				nil,
			)
		}
		return nil
	}
}

// ValidateSingle passes if at least one arg is specified
func ValidateSingle(argName string) clix.ValidateFunc {
	return func(args []string) error {
		if len(args) < 1 {
			return errors.E(
				errors.Invalid,
				fmt.Sprintf("required argument %s not found", argName),
				nil,
			)
		}
		return nil
	}
}
