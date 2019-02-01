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

// DevicesRepository provides methods to manage Devices
type DevicesRepository struct {
	logger log.FieldLogger
	client coreclient.DevicesGetter

	// cache holds a mapping from namespace to map from name to Device
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]*corev1beta1.Device

	// pubKeyCache caches the devices by public key
	// it is protected by a mutex
	pubKeyMutex sync.RWMutex
	pubKeyCache map[string]*corev1beta1.Device

	// serialnumberCache caches the devices by serialnumbers
	// it is protected by a mutex
	serialnumberMutex sync.RWMutex
	serialnumberCache map[string]*corev1beta1.Device
}

// NewDevicesRepository creates a new Devices repository with the given
// settings
func NewDevicesRepository(
	logger log.FieldLogger,
	client coreclient.DevicesGetter,
	informer Informer,
	resync time.Duration,
) (*DevicesRepository, error) {
	r := &DevicesRepository{
		logger:            logger,
		client:            client,
		cache:             make(map[string]map[string]*corev1beta1.Device),
		pubKeyCache:       make(map[string]*corev1beta1.Device),
		serialnumberCache: make(map[string]*corev1beta1.Device),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new Device is added
func (d *DevicesRepository) OnAdd(obj interface{}) {
	device, ok := obj.(*corev1beta1.Device)
	if !ok {
		d.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	if _, ok := d.cache[device.Namespace]; !ok {
		d.cache[device.Namespace] = make(map[string]*corev1beta1.Device)
	}

	d.cache[device.Namespace][device.Name] = device
	d.mutex.Unlock()

	if device.Spec.PublicKey != nil {
		d.pubKeyMutex.Lock()
		d.pubKeyCache[*device.Spec.PublicKey] = device
		d.pubKeyMutex.Unlock()
	}
	if device.Spec.Serialnumber != "" {
		d.serialnumberMutex.Lock()
		d.serialnumberCache[device.Spec.Serialnumber] = device
		d.serialnumberMutex.Unlock()
	}
}

// OnUpdate is called by the informer when an existing Device is updated
func (d *DevicesRepository) OnUpdate(oldObj, newObj interface{}) {
	device, ok := newObj.(*corev1beta1.Device)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	d.mutex.Lock()

	d.cache[device.Namespace][device.Name] = device
	d.mutex.Unlock()

	if device.Spec.PublicKey != nil {
		d.pubKeyMutex.Lock()
		d.pubKeyCache[*device.Spec.PublicKey] = device
		d.pubKeyMutex.Unlock()
	}
	if device.Spec.Serialnumber != "" {
		d.serialnumberMutex.Lock()
		d.serialnumberCache[device.Spec.Serialnumber] = device
		d.serialnumberMutex.Unlock()
	}
}

// OnDelete is called by the informer when a Device is deleted
func (d *DevicesRepository) OnDelete(obj interface{}) {
	device, ok := obj.(*corev1beta1.Device)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	delete(d.cache[device.Namespace], device.Name)
	d.mutex.Unlock()

	if device.Spec.PublicKey != nil {
		d.pubKeyMutex.Lock()
		delete(d.pubKeyCache, *device.Spec.PublicKey)
		d.pubKeyMutex.Unlock()
	}
	if device.Spec.Serialnumber != "" {
		d.serialnumberMutex.Lock()
		delete(d.serialnumberCache, device.Spec.Serialnumber)
		d.serialnumberMutex.Unlock()
	}
}

// Create creates a new Device
func (d *DevicesRepository) Create(ctx context.Context, device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return d.client.Devices(device.Namespace).Create(device)
}

// Update updates an existing Device
func (d *DevicesRepository) Update(ctx context.Context, device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return d.client.Devices(device.Namespace).Update(device)
}

// UpdateStatus updates the status of an an existing Device
func (d *DevicesRepository) UpdateStatus(ctx context.Context, device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return d.client.Devices(device.Namespace).UpdateStatus(device)
}

// Get gets the Device with the given name in the given namespace
func (d *DevicesRepository) Get(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if device, ok := d.cache[namespace][name]; ok {
		return device, nil
	}

	return nil, fmt.Errorf("not found")
}

// GetByPublicKey returns the device with the given public Key
func (d *DevicesRepository) GetByPublicKey(ctx context.Context, publicKey string) (*corev1beta1.Device, error) {
	d.pubKeyMutex.RLock()
	defer d.pubKeyMutex.RUnlock()

	if device, ok := d.pubKeyCache[publicKey]; ok {
		return device, nil
	}

	return nil, fmt.Errorf("not found")
}

// GetBySerialnumber returns the device with the given public Key
func (d *DevicesRepository) GetBySerialnumber(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
	d.serialnumberMutex.RLock()
	defer d.serialnumberMutex.RUnlock()

	if device, ok := d.serialnumberCache[serialnumber]; ok {
		return device, nil
	}

	return nil, fmt.Errorf("not found")
}

// List returns the list of Devices in the given namespace
func (d *DevicesRepository) List(ctx context.Context, namespace string) ([]*corev1beta1.Device, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if devices, ok := d.cache[namespace]; ok {
		var result []*corev1beta1.Device
		for _, device := range devices {
			result = append(result, device)
		}
		return result, nil
	}

	return []*corev1beta1.Device{}, nil
}

// Delete deletes the Device with the given name in the given namespace
func (d *DevicesRepository) Delete(ctx context.Context, namespace, name string) error {
	return d.client.Devices(namespace).Delete(name, nil)
}
