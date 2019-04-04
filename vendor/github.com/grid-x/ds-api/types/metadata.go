package types

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Metadata contains metadata for a all resources
type Metadata struct {
	ID          string            `json:"id"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// UpdateMetadata contains metadata which can be updated
type UpdateMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// ConvertFromK8sMetadata covernts k8s metadata to internal metadata
func ConvertFromK8sMetadata(meta metav1.ObjectMeta, unfiltered bool) Metadata {
	if unfiltered {
		return Metadata{
			ID:          meta.Name,
			Labels:      meta.Labels,
			Annotations: meta.Annotations,
		}
	}

	return Metadata{
		ID:          meta.Name,
		Labels:      meta.Labels,
		Annotations: filterAnnotations(meta.Annotations),
	}
}

func filterAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	for k := range annotations {
		if !strings.HasPrefix(k, "gridx.ai") {
			delete(annotations, k)
		}
	}

	return annotations
}
