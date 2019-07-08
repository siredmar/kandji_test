package api

import (
	"fmt"
	"strings"
)

// ParseMetadataMap is used to setup a sting map based on an input string of format "a=b c=d..."
func ParseMetadataMap(input string) (map[string]string, error) {
	labels := strings.Split(input, " ")
	m := make(map[string]string)
	for _, pair := range labels {
		z := strings.Split(pair, "=")
		if len(z) > 2 {
			return nil, fmt.Errorf("Wrong metdata format. Expected 'a=b c=d...'")
		}
		if len(z) == 1 {
			if !strings.HasSuffix(z[0], "-") {
				return nil, fmt.Errorf("Wrong metdata format. 'abc-' to remove an entry")
			}
		}
		if len(z) == 1 {
			m[z[0]] = ""
		}
		if len(z) == 2 {
			m[z[0]] = z[1]
		}
	}
	return m, nil
}
