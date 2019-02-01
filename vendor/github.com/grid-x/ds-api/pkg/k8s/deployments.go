package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	appsclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/apps/v1beta1"
	log "github.com/sirupsen/logrus"
)

// DeploymentsRepository provides methods to manage deployments
type DeploymentsRepository struct {
	logger log.FieldLogger
	client appsclient.DeviceDeploymentsGetter

	// cache holds a mapping from namespace to map from name to deployment
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]*appsv1beta1.DeviceDeployment
}

// NewDeploymentsRepository creates a new deployments repository with the given
// settings
func NewDeploymentsRepository(
	logger log.FieldLogger,
	client appsclient.DeviceDeploymentsGetter,
	informer Informer,
	resync time.Duration,
) (*DeploymentsRepository, error) {
	r := &DeploymentsRepository{
		logger: logger,
		client: client,
		cache:  make(map[string]map[string]*appsv1beta1.DeviceDeployment),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new deployment is added
func (d *DeploymentsRepository) OnAdd(obj interface{}) {
	deploy, ok := obj.(*appsv1beta1.DeviceDeployment)
	if !ok {
		d.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.cache[deploy.Namespace]; !ok {
		d.cache[deploy.Namespace] = make(map[string]*appsv1beta1.DeviceDeployment)
	}

	d.cache[deploy.Namespace][deploy.Name] = deploy
}

// OnUpdate is called by the informer when an existing deployment is updated
func (d *DeploymentsRepository) OnUpdate(oldObj, newObj interface{}) {
	deploy, ok := newObj.(*appsv1beta1.DeviceDeployment)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[deploy.Namespace][deploy.Name] = deploy
}

// OnDelete is called by the informer when a deployment is deleted
func (d *DeploymentsRepository) OnDelete(obj interface{}) {
	deploy, ok := obj.(*appsv1beta1.DeviceDeployment)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.cache[deploy.Namespace], deploy.Name)
}

// Create creates a new deployment
func (d *DeploymentsRepository) Create(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return d.client.DeviceDeployments(deploy.Namespace).Create(deploy)
}

// Update updates an existing deployment
func (d *DeploymentsRepository) Update(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return d.client.DeviceDeployments(deploy.Namespace).Update(deploy)
}

// Get gets the deployment with the given name in the given namespace
func (d *DeploymentsRepository) Get(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceDeployment, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if deploy, ok := d.cache[namespace][name]; ok {
		return deploy, nil
	}

	return nil, fmt.Errorf("not found")
}

// List returns the list of deployments in the given namespace
func (d *DeploymentsRepository) List(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceDeployment, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if deploys, ok := d.cache[namespace]; ok {
		var result []*appsv1beta1.DeviceDeployment
		for _, deploy := range deploys {
			result = append(result, deploy)
		}
		return result, nil
	}

	return []*appsv1beta1.DeviceDeployment{}, nil
}

// Delete deletes the deployment with the given name in the given namespace
func (d *DeploymentsRepository) Delete(ctx context.Context, namespace, name string) error {
	return d.client.DeviceDeployments(namespace).Delete(name, nil)
}
