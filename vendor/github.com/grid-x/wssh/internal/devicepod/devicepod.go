package devicepod

import (
	"context"
	"fmt"

	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/wssh/internal"
	"github.com/grid-x/wssh/internal/k8s"
	"github.com/grid-x/wssh/internal/model"
	"github.com/grid-x/wssh/pkg/tunnel"
)

const (
	wsshTag = "core.gridx.ai/wssh"
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
	logLevel string,
	keysPath string,
	compressLogs bool,
) (*corev1beta1.DevicePod, error) {
	log = log.WithField("routine", "CreatePod")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podtemplate := getSSHPodTemplate(sessionCtx, deviceImage, externalAddr, dsAddr, tunnel.IDString(deviceTunnel.ID), logLevel, keysPath, compressLogs)
	pod, err := pods.Create(ctx, podtemplate)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// DeletePod deletes a pod for a specific tunnel
func DeletePod(log logrus.FieldLogger, podRepo *k8s.PodsRepository, accountID, deviceID, tunnelID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pod, err := podRepo.Get(ctx, model.AccountNamespaceName(accountID), tunnelID)
	if err != nil {
		return err
	}

	if _, ok := pod.ObjectMeta.Annotations[wsshTag]; !ok {
		return fmt.Errorf(fmt.Sprintf("Pod %s did not contain wsshTag", pod.Name))
	}
	if pod.Spec.DeviceID != deviceID {
		return fmt.Errorf(fmt.Sprintf("Pod %s was assigned to a different device, want %s, got %s", pod.Name, deviceID, pod.Spec.DeviceID))
	}

	err = podRepo.Delete(ctx, pod)
	if err != nil {
		return err
	}
	log.WithField("devicePod", pod.Name).Info("delete")

	return nil
}

func getSSHPodTemplate(sessionCtx internal.SessionContext, image, externalAddr, dsAddr, tID, logLevel, keysPath string, compressLogs bool) *corev1beta1.DevicePod {
	dir := corev1beta1.HostPathDirectory
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tID,
			Namespace: model.AccountNamespaceName(sessionCtx.AccountID),
			Annotations: map[string]string{
				wsshTag: "true",
			},
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: sessionCtx.DeviceID,
			Config: corev1beta1.PodConfig{
				Volumes: []corev1beta1.Volume{
					{
						Name: "keys-host",
						HostPath: &corev1beta1.HostPathVolumeSource{
							Path: keysPath,
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
					{
						Name: "tmp",
						HostPath: &corev1beta1.HostPathVolumeSource{
							Path: "/tmp",
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
								Name:  "LOG_LEVEL",
								Value: logLevel,
							},
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
							{
								Name:  "SID",
								Value: sessionCtx.SessionID,
							},
							{
								Name:  "COMPRESS_LOGS",
								Value: fmt.Sprintf("%v", compressLogs),
							},
						},
						VolumeMounts: []corev1beta1.VolumeMount{
							{
								Name:      "keys-host",
								MountPath: "/keys-host",
							},
							{
								Name:      "runtime",
								ReadOnly:  true,
								MountPath: "/var/run",
							},
							{
								Name:      "tmp",
								MountPath: "/tmp",
							},
						},
					},
				},
			},
		},
	}
}
