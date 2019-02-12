package api

import (
	"errors"
	"fmt"
	"strings"
)

func LookupID(prefix string, ids []string) (string, error) {
	if len(prefix) == 36 {
		//Not an prefix at all but the full ID
		return prefix, nil
	}

	var out string

	for _, i := range ids {
		if strings.HasPrefix(i, prefix) {
			if out != "" {
				s := fmt.Sprintf("Found more then one result for abbreviation %s", prefix)
				return out, errors.New(s)
			}
			out = i
		}
	}

	if out == "" {
		s := fmt.Sprintf("Found no result for abbreviation %s", prefix)
		return out, errors.New(s)
	}

	return out, nil
}
