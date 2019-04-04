package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Annotations to filter out
const (
	InternalAnnotation = "core.gridx.ai/ds-system"
)

// IsAllowed checks if a resource should be filtered out as it's internal
func IsAllowed(meta metav1.ObjectMeta, unfiltered bool) bool {
	if unfiltered {
		return true
	}

	// Return false to dissallow if it has internal annotations
	return !hasInternalAnnotation(meta.Annotations)
}

func hasInternalAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}

	for k, v := range annotations {
		// Disallow elements including system annotation => core.gridx.ai/ds-system: true
		if k == InternalAnnotation && v == "true" {
			return true
		}
	}

	return false
}
