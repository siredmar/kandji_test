package k8s

import (
	"context"

	corev1beta1 "github.com/grid-x/ds-k8s-v2/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"
	types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DeviceIDIndexField is an index name for caching devices by their id
	DeviceIDIndexField = ".index.deviceID"
)

// PodsRepository provides methods to manage pods
type PodsRepository struct {
	logger log.FieldLogger
	client client.Client
}

// NewPodsRepository creates a new pods repository with the given
// settings
func NewPodsRepository(
	logger log.FieldLogger,
	client client.Client,
) (*PodsRepository, error) {
	r := &PodsRepository{
		logger: logger,
		client: client,
	}

	return r, nil
}

// UpdateStatus updates the status of an existing pod
func (p *PodsRepository) UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	err := p.client.Status().Update(ctx, pod)
	return pod, err
}

// Get gets the pod with the given name in the given namespace
func (p *PodsRepository) Get(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
	var pod corev1beta1.DevicePod
	err := p.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &pod)
	return &pod, err
}

// Create creates a new pod
func (p *PodsRepository) Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	err := p.client.Create(ctx, pod)
	return pod, err
}

// Delete deletes the pod with the given name in the given namespace
func (p *PodsRepository) Delete(ctx context.Context, pod *corev1beta1.DevicePod) error {
	return p.client.Delete(ctx, pod)
}

// List returns the list of pods in the given namespace
func (p *PodsRepository) List(ctx context.Context, namespace string) ([]corev1beta1.DevicePod, error) {
	var pods corev1beta1.DevicePodList
	if err := p.client.List(ctx, &pods, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// ListByDeviceID returns the pods scheduled for a specific device with ID
// deviceID in the given namespace
func (p *PodsRepository) ListByDeviceID(ctx context.Context, namespace, deviceID string) ([]corev1beta1.DevicePod, error) {
	var pods corev1beta1.DevicePodList
	if err := p.client.List(ctx, &pods, client.InNamespace(namespace), client.MatchingFields{DeviceIDIndexField: deviceID}); err != nil {
		return nil, err
	}

	return pods.Items, nil
}
