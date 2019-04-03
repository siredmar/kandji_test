package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
	configclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/config/v1beta1"
	log "github.com/sirupsen/logrus"
)

// DockerConfigRepository implements methods to manage DockerConfigs
type DockerConfigRepository struct {
	logger log.FieldLogger
	client configclient.DockerConfigsGetter

	// cache holds a mapping from namespace to map from name to dockerConfig
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]*configv1beta1.DockerConfig
}

// NewDockerConfigRepository creates a new dockerConfig repository with the
// given settings
func NewDockerConfigRepository(
	logger log.FieldLogger,
	client configclient.DockerConfigsGetter,
	informer Informer,
	resync time.Duration,
) (*DockerConfigRepository, error) {
	r := &DockerConfigRepository{
		logger: logger,
		client: client,
		cache:  make(map[string]map[string]*configv1beta1.DockerConfig),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new application is added
func (cr *DockerConfigRepository) OnAdd(obj interface{}) {
	c, ok := obj.(*configv1beta1.DockerConfig)
	if !ok {
		cr.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if _, ok := cr.cache[c.Namespace]; !ok {
		cr.cache[c.Namespace] = make(map[string]*configv1beta1.DockerConfig)
	}

	cr.cache[c.Namespace][c.Name] = c
}

// OnUpdate is called by the informer when an application is updated
func (cr *DockerConfigRepository) OnUpdate(oldObj, newObj interface{}) {
	c, ok := newObj.(*configv1beta1.DockerConfig)
	if !ok {
		cr.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	cr.cache[c.Namespace][c.Name] = c
}

// OnDelete is called by the informer when an application is deleted
func (cr *DockerConfigRepository) OnDelete(obj interface{}) {
	c, ok := obj.(*configv1beta1.DockerConfig)
	if !ok {
		cr.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	delete(cr.cache[c.Namespace], c.Name)
}

// Create creates a new application
func (cr *DockerConfigRepository) Create(ctx context.Context, config *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
	return cr.client.DockerConfigs(config.Namespace).Create(config)
}

// Get returns the dockerConfig with the given name if found
func (cr *DockerConfigRepository) Get(ctx context.Context, namespace, name string) (*configv1beta1.DockerConfig, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	if c, ok := cr.cache[namespace][name]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("not found")
}

// List lists all dockerConfigs in a given namespace
func (cr *DockerConfigRepository) List(ctx context.Context, namespace string) ([]*configv1beta1.DockerConfig, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	if configs, ok := cr.cache[namespace]; ok {
		var result []*configv1beta1.DockerConfig
		for _, c := range configs {
			result = append(result, c)
		}
		return result, nil
	}

	return []*configv1beta1.DockerConfig{}, nil
}

// Delete deletes the application with the given name
func (cr *DockerConfigRepository) Delete(ctx context.Context, namespace, name string) error {
	return cr.client.DockerConfigs(namespace).Delete(name, nil)
}
