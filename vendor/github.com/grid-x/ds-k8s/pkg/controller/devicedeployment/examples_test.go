package devicedeployment

import (
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

const (
	gridXdeNamespace = "gridx-de"
)

var (
	stableChannelLabel = map[string]string{
		GridXChannelLabel: GridXStableChannel,
	}
	alphaChannelLabel = map[string]string{
		GridXChannelLabel: "alpha",
	}
	vendorSolaredgeLabel = map[string]string{
		"vendor": "solaredge",
	}

	// Device examples
	gridBox001Stable          = mkDevice("gridbox001", stableChannelLabel)
	gridBox001StableSolaredge = mkDevice(
		"gridbox001",
		combineLabels(
			stableChannelLabel,
			vendorSolaredgeLabel,
		),
	)
	gridBox001Alpha  = mkDevice("gridbox001", alphaChannelLabel)
	gridBox002Stable = mkDevice("gridbox002", stableChannelLabel)

	// DevicePod examples

	// DeviceDeployment examples
	envscanStableDeployment          = mkDeployment("envscan-stable", "envscan", nil, stableChannelLabel)
	envscanDevice001Deployment       = mkDeployment("envscan-gridbox001", "envscan", mkString("gridbox001"), nil)
	envscanDevice002Deployment       = mkDeployment("envscan-gridbox002", "envscan", mkString("gridbox002"), nil)
	envscanAlphaDeployment           = mkDeployment("envscan-alpha", "envscan", nil, alphaChannelLabel)
	envscanAlphaDevice001Deployment  = mkDeployment("envscan-alpha", "envscan", mkString("gridbox001"), alphaChannelLabel)
	envscanStableSolaredgeDeployment = mkDeployment(
		"envscan-stable-solaredge",
		"envscan",
		nil,
		combineLabels(stableChannelLabel, vendorSolaredgeLabel),
	)
	monitoringStableDeployment    = mkDeployment("monitoring-stable", "monitoring", nil, stableChannelLabel)
	monitoringDevice001Deployment = mkDeployment("monitoring-gridbox001", "monitoring", mkString("gridbox001"), nil)
)

func mkDevice(name string, labels map[string]string) *corev1beta1.Device {
	return &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: gridXdeNamespace,
			Labels:    labels,
			UID:       types.UID(uuid.New().String()),
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}
}

func mkDeployment(name, app string, deviceID *string, labels map[string]string) *appsv1beta1.DeviceDeployment {
	d := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: gridXdeNamespace,
			UID:       types.UID(uuid.New().String()),
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App:      app,
			Selector: appsv1beta1.Selector{},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Image:   "gridx.de/" + app + ":stable",
							Command: []string{"/usr/local/bin/" + app},
						},
					},
					RestartPolicy: &corev1beta1.RestartPolicy{
						Type: corev1beta1.RestartPolicyAlways,
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	if labels != nil {
		d.Spec.Selector.MatchByLabels = labels
	}
	if deviceID != nil {
		d.Spec.Selector.MatchByDeviceID = deviceID
	}

	return d
}

func mkContainer(deploy *appsv1beta1.DeviceDeployment, deviceID string) *corev1beta1.DevicePod {
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: gridXdeNamespace,
			Annotations: map[string]string{
				appsv1beta1.AppNameAnnotation: deploy.Spec.App,
			},
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: deviceID,
			Config:   deploy.Spec.Template.Spec,
		},
		Status: corev1beta1.DevicePodStatus{},
	}
}

// combineLabels combines a list of maps, i.e. list of labels to a single map
// In this process duplicate keys get the last value present in the list
func combineLabels(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		if m == nil {
			continue
		}

		// Overwrite any existing values
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func withLabels(obj interface{}, labels map[string]string) interface{} {
	switch o := obj.(type) {
	case *corev1beta1.Device:
		box := o.DeepCopy()
		box.Labels = combineLabels(box.Labels, labels)
		return box
	case *appsv1beta1.DeviceDeployment:
		d := o.DeepCopy()
		d.Spec.Selector.MatchByLabels = combineLabels(
			d.Spec.Selector.MatchByLabels,
			labels,
		)
		return d
	default:
		return nil
	}
}

func withSelector(d *appsv1beta1.DeviceDeployment, s appsv1beta1.Selector) *appsv1beta1.DeviceDeployment {
	copy := d.DeepCopy()
	copy.Spec.Selector = s
	return copy
}

func withApp(d *appsv1beta1.DeviceDeployment, app string) *appsv1beta1.DeviceDeployment {
	copy := d.DeepCopy()
	copy.Spec.App = app
	return copy
}

func withPodConfig(d *appsv1beta1.DeviceDeployment, podConfig corev1beta1.PodConfig) *appsv1beta1.DeviceDeployment {
	copy := d.DeepCopy()
	copy.Spec.Template.Spec = podConfig
	return copy
}

func withStatus(d *appsv1beta1.DeviceDeployment, st appsv1beta1.DeviceDeploymentStatus) *appsv1beta1.DeviceDeployment {
	copy := d.DeepCopy()
	copy.Status = st
	return copy
}
