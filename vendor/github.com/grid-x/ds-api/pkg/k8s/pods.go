package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	coreclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/core/v1beta1"
	log "github.com/sirupsen/logrus"
)

// PodsRepository provides methods to manage pods
type PodsRepository struct {
	logger log.FieldLogger
	client coreclient.DevicePodsGetter

	// cache holds a mapping from namespace to map from deviceID to a map
	// from name of the pod to the pod itself
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]map[string]*corev1beta1.DevicePod
}

// NewPodsRepository creates a new pods repository with the given
// settings
func NewPodsRepository(
	logger log.FieldLogger,
	client coreclient.DevicePodsGetter,
	informer Informer,
	resync time.Duration,
) (*PodsRepository, error) {
	r := &PodsRepository{
		logger: logger,
		client: client,
		cache:  make(map[string]map[string]map[string]*corev1beta1.DevicePod),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new pod is added
func (p *PodsRepository) OnAdd(obj interface{}) {
	pod, ok := obj.(*corev1beta1.DevicePod)
	if !ok {
		p.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.cache[pod.Namespace]; !ok {
		p.cache[pod.Namespace] = make(map[string]map[string]*corev1beta1.DevicePod)
	}

	if _, ok := p.cache[pod.Namespace][pod.Spec.DeviceID]; !ok {
		p.cache[pod.Namespace][pod.Spec.DeviceID] = make(map[string]*corev1beta1.DevicePod)
	}

	p.cache[pod.Namespace][pod.Spec.DeviceID][pod.Name] = pod
}

// OnUpdate is called by the informer when an existing pod is updated
func (p *PodsRepository) OnUpdate(oldObj, newObj interface{}) {
	pod, ok := newObj.(*corev1beta1.DevicePod)
	if !ok {
		p.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.cache[pod.Namespace][pod.Spec.DeviceID][pod.Name] = pod
}

// OnDelete is called by the informer when a pod is deleted
func (p *PodsRepository) OnDelete(obj interface{}) {
	pod, ok := obj.(*corev1beta1.DevicePod)
	if !ok {
		p.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.cache[pod.Namespace][pod.Spec.DeviceID], pod.Name)
}

// UpdateStatus updates the status of an existing pod
func (p *PodsRepository) UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return p.client.DevicePods(pod.Namespace).UpdateStatus(pod)
}

// Get gets the pod with the given name in the given namespace
func (p *PodsRepository) Get(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for dev := range p.cache[namespace] {
		if pod, ok := p.cache[namespace][dev][name]; ok {
			return pod, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

// Create creates a new pod
func (p *PodsRepository) Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return p.client.DevicePods(pod.Namespace).Create(pod)
}

// Delete deletes the pod with the given name in the given namespace
func (p *PodsRepository) Delete(ctx context.Context, namespace, name string) error {
	return p.client.DevicePods(namespace).Delete(name, nil)
}

// List returns the list of pods in the given namespace
func (p *PodsRepository) List(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if devices, ok := p.cache[namespace]; ok {
		var result []*corev1beta1.DevicePod
		for _, pods := range devices {
			for _, pod := range pods {
				result = append(result, pod)
			}
		}
		return result, nil
	}

	return []*corev1beta1.DevicePod{}, nil
}

// ListByDeviceID returns the pods scheduled for a specific device with ID
// deviceID in the given namespace
func (p *PodsRepository) ListByDeviceID(ctx context.Context, namespace, deviceID string) ([]*corev1beta1.DevicePod, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if _, ok := p.cache[namespace]; ok {
		if _, ok := p.cache[namespace][deviceID]; ok {
			var result []*corev1beta1.DevicePod
			for _, pod := range p.cache[namespace][deviceID] {
				result = append(result, pod)
			}
			return result, nil
		}
	}

	return []*corev1beta1.DevicePod{}, nil
}
