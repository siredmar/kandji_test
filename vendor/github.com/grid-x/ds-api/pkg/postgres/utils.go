package postgres

import (
	"fmt"
	"strings"
)

// prepareOrderQuery adjust the a query by replacing $sort and $direction.
// It returns the identifier for the query = $sort:$direction and the prepared
// query.
func prepareOrderQuery(q, sort, direction string) (string, string) {
	r := strings.NewReplacer("$sort", sort, "$direction", direction)
	return fmt.Sprintf("%s:%s", sort, direction), r.Replace(q)
}

// validOrDefault returns a string if the input is in a list or
// returns the specified default.
func validOrDefault(input string, list []string, def string) string {
	for _, l := range list {
		if strings.EqualFold(input, l) {
			return l
		}
	}
	return def
}
