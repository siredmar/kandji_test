package devicepod

import (
	"context"

	"github.com/google/uuid"
	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/wssh/internal"
	"github.com/grid-x/wssh/internal/k8s"
	"github.com/grid-x/wssh/internal/model"
	"github.com/grid-x/wssh/pkg/tunnel"
)

const (
	wsshTag    = "core.gridx.ai/wssh"
	wsshTagTID = "core.gridx.ai/wssh-tid"
)

// Create a new DevicePod
func Create(
	log logrus.FieldLogger,
	pods *k8s.PodsRepository,
	sessionCtx internal.SessionContext,
	deviceTunnel *tunnel.Tunnel,
	deviceImage string,
	externalAddr string,
	dsAddr string,
) (*corev1beta1.DevicePod, error) {
	log = log.WithField("routine", "CreatePod")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podtemplate := getSSHPodTemplate(sessionCtx, deviceImage, externalAddr, dsAddr, tunnel.IDString(deviceTunnel.ID))
	pod, err := pods.Create(ctx, podtemplate)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// DeleteAll existing DevicePods for a device
func DeleteAll(log logrus.FieldLogger, podRepo *k8s.PodsRepository, accountID, deviceID, tunnelID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pods, err := podRepo.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return err
	}

	for _, p := range pods {
		if _, ok := p.ObjectMeta.Annotations[wsshTag]; !ok {
			continue
		}
		if tID, ok := p.ObjectMeta.Annotations[wsshTagTID]; !ok || tID != tunnelID {
			continue
		}
		if p.Spec.DeviceID != deviceID {
			continue
		}

		err := podRepo.Delete(ctx, &p)
		if err != nil {
			return err
		}
		log.WithField("devicePod", p.Name).Info("delete")
	}

	return nil
}

func getSSHPodTemplate(sessionCtx internal.SessionContext, image, externalAddr, dsAddr, tID string) *corev1beta1.DevicePod {
	dir := corev1beta1.HostPathDirectory
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: model.AccountNamespaceName(sessionCtx.AccountID),
			Annotations: map[string]string{
				wsshTag:    "true",
				wsshTagTID: tID,
			},
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: sessionCtx.DeviceID,
			Config: corev1beta1.PodConfig{
				Volumes: []corev1beta1.Volume{
					{
						Name: "keys",
						HostPath: &corev1beta1.HostPathVolumeSource{
							Path: "/mnt/state/root-overlay/etc/dropbear",
							Type: &dir,
						},
					},
					{
						Name: "runtime",
						HostPath: &corev1beta1.HostPathVolumeSource{
							Path: "/var/run",
							Type: &dir,
						},
					},
				},
				Network: "Host",
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
								Name:  "DS_ADDR",
								Value: dsAddr,
							},
							{
								Name:  "TID",
								Value: tID,
							},
						},
						VolumeMounts: []corev1beta1.VolumeMount{
							{
								Name:      "keys",
								ReadOnly:  true,
								MountPath: "/keys",
							},
							{
								Name:      "runtime",
								ReadOnly:  true,
								MountPath: "/var/run",
							},
						},
					},
				},
			},
		},
	}
}
