package e2e

import (
	"strings"

	"github.com/google/go-cmp/cmp"
)

func IgnoreSuffixes(suffixes ...string) cmp.Option {
	return cmp.FilterPath(
		func(p cmp.Path) bool {
			for _, suffix := range suffixes {
				if strings.HasSuffix(p.String(), suffix) {
					return true
				}
			}
			return false
		},
		cmp.Ignore(),
	)
}
