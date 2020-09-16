package devicepod

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/wssh/internal/k8s"
	"github.com/grid-x/wssh/internal/model"
	"github.com/grid-x/wssh/pkg/tunnel"
)

const wsshTag = "core.gridx.ai/wssh"

// Create a new DevicePod
func Create(
	log logrus.FieldLogger,
	pods *k8s.PodsRepository,
	accountID string,
	deviceID string,
	deviceTunnel *tunnel.Tunnel,
	deviceImage string,
	externalAddr string,
) (*corev1beta1.DevicePod, error) {
	log = log.WithField("routine", "CreatePod")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podtemplate := getSSHPodTemplate(deviceID, accountID, deviceImage, externalAddr, fmt.Sprintf("%v", deviceTunnel.ID))
	pod, err := pods.Create(ctx, podtemplate)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// Delete an existing DevicePod
func Delete(log logrus.FieldLogger) {}

func getSSHPodTemplate(deviceID, accountID, image, externalAddr, tID string) *corev1beta1.DevicePod {
	dir := corev1beta1.HostPathDirectory
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        uuid.New().String(),
			Namespace:   model.AccountNamespaceName(accountID),
			Annotations: map[string]string{wsshTag: "true"},
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: deviceID,
			Config: corev1beta1.PodConfig{
				Volumes: []corev1beta1.Volume{
					{
						Name: "runtime",
						HostPath: &corev1beta1.HostPathVolumeSource{
							Path: "/var/run",
							Type: &dir,
						},
					},
				},
				Containers: []corev1beta1.Container{
					{
						Name:  "wssh",
						Image: image,
						Environment: []corev1beta1.EnvVar{
							{
								Name:  "SERVER_ADDR",
								Value: externalAddr,
							},
							{
								Name:  "TID",
								Value: tID,
							},
						},
						VolumeMounts: []corev1beta1.VolumeMount{
							{
								Name: "runtime",
								MountPath: "/var/run",
							},
						},
					},
				},
			},
		},
	}
}
