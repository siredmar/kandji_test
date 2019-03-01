package model

import (
	"strings"
)

// ComputeLabels is merging label maps
func ComputeLabels(current map[string]string, update map[string]string) map[string]string {
	if update == nil {
		return current
	}
	if current == nil {
		current = make(map[string]string)
	}

	for k, v := range update {
		if strings.HasSuffix(k, "-") {
			// Call to delete
			delete(current, k[:len(k)-1])
		} else {
			current[k] = v
		}
	}

	return current
}
