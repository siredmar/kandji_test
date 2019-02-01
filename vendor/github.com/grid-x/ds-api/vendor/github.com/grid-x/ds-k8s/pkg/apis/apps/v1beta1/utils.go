package v1beta1

import (
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

const (
	// AppNameAnnotation is the app name of the annotation that stores the app name
	AppNameAnnotation = "gridx.ai/app"
)

// ExtractAppName extracts the AppName from a GridBoxContainer resource
func ExtractAppName(obj corev1beta1.DevicePod) string {
	return obj.ObjectMeta.Annotations[AppNameAnnotation]
}
