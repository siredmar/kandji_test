package types

import (
	"strings"
)

// Metadata contains metadata for a all resources
type Metadata struct {
	ID          string            `json:"id"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// UpdateMetadata contains metadata which can be updated
type UpdateMetadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func filterAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	for k := range annotations {
		if !strings.Contains(k, "gridx.ai") {
			delete(annotations, k)
		}
	}

	return annotations
}
