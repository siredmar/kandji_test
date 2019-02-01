package testing

import (
	"strings"

	"github.com/google/go-cmp/cmp"
)

// CmpWithIgnore will add a custom comparer for string fields whose path ends
// with the given path.
// This comparer will ignore this filed if either of the two fieds is set to
// "@ignore". Otherwise it will compare the fields as usual.
func CmpWithIgnore(path string) cmp.Option {
	return cmp.FilterPath(
		func(p cmp.Path) bool {
			return strings.HasSuffix(p.String(), path)
		},
		cmp.Comparer(func(x, y string) bool {
			if x == "" && y == "" {
				return true
			} else if x == "@ignore" || y == "@ignore" {
				return true
			} else {
				return x == y
			}
		}))
}
