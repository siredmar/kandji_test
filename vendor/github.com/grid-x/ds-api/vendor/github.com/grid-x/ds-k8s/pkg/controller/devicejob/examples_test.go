package devicejob

import (
	"github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

const (
	gridXdeNamespace = "gridx-de"
)

var (
	stableChannelLabel = map[string]string{
		"gridx.de/channel": "stable",
	}
	alphaChannelLabel = map[string]string{
		"gridx.de/channel": "alpha",
	}

	// Device examples
	gridBox001Stable = mkDevice("gridbox001", stableChannelLabel)
	gridBox001Alpha  = mkDevice("gridbox001", alphaChannelLabel)
	gridBox002Stable = mkDevice("gridbox002", stableChannelLabel)
)

func mkDevice(name string, labels map[string]string) *corev1beta1.Device {
	return &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: gridXdeNamespace,
			Labels:    labels,
			UID:       types.UID(uuid.NewV4().String()),
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}
}
