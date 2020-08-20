package k8s

import (
	"context"

	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DevicesRepository provides methods to manage Devices
type DevicesRepository struct {
	logger log.FieldLogger
	client client.Client
}

// NewDevicesRepository creates a new Devices repository with the given
// settings
func NewDevicesRepository(
	logger log.FieldLogger,
	client client.Client,
) (*DevicesRepository, error) {
	r := &DevicesRepository{
		logger: logger,
		client: client,
	}

	return r, nil
}

// Get gets the Device with the given name in the given namespace
func (d *DevicesRepository) Get(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
	dev := &corev1beta1.Device{}
	err := d.client.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, dev)

	return dev, err
}
