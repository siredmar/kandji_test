package action

import (
	"fmt"

	"github.com/grid-x/gxctl/internal/lint"
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Lint(s *service.Service, lintCmdFilename string, quiet bool) error {
	// read raw file contents into fileName -> bytes map
	// and parse file contents into fileName -> resource map
	contents, err := api.GetFilesContentsToProcess(lintCmdFilename)
	if err != nil {
		return err
	}

	var fsNames []string
	fsResources := make(map[string]interface{})

	for n, c := range contents {
		res, _, err := api.CheckResourceFile(c, false)
		if err != nil {
			return err
		}
		fsNames = append(fsNames, n)
		fsResources[n] = res
	}

	// init context with current state
	ctx, err := context.New(s.Client)
	if err != nil {
		return err
	}

	// build permutations:
	// for each file, we need a context which includes every other file but not itself
	jobs := make(map[string][]string)
	for i, x := range fsNames {
		tmp := make([]string, len(fsNames))
		copy(tmp, fsNames)
		jobs[x] = append(tmp[:i], tmp[i+1:]...)
	}

	results := []result.Result{}
	for fileName, context := range jobs {
		desired := []interface{}{}

		for _, ctxRes := range context {
			desired = append(desired, fsResources[ctxRes])
		}
		ctx.SetDesired(desired)

		result, err := lint.Lint(ctx, fileName, fsResources[fileName])
		if err != nil {
			fmt.Println(err)
		}
		results = append(results, result...)
	}

	var total, pass, skip int
	for _, result := range results {
		if result.Skip {
			skip++
			continue
		}
		total++
		if result.Pass {
			pass++
		}
	}

	if !quiet {
		fmt.Printf("%v / %v passed (%v skipped)\n", pass, total, skip)
		for _, result := range results {
			fmt.Println(result)
		}
	}

	if pass != total {
		var errs []string

		if quiet {
			for _, r := range results {
				if !r.Pass {
					errs = append(errs, r.String())
				}
			}
		}
		return errors.E(
			errors.Validation,
			"Some linting rules did not pass",
			errs,
		)
	}

	return nil
}
