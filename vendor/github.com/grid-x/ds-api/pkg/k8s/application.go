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

// ApplicationsRepository implements methods to manage Applications
type ApplicationsRepository struct {
	logger log.FieldLogger
	client appsclient.DeviceApplicationsGetter

	// cache holds a mapping from namespace to map from name to application
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]*appsv1beta1.DeviceApplication
}

// NewApplicationsRepository creates a new applications repository with the
// given settings
func NewApplicationsRepository(
	logger log.FieldLogger,
	client appsclient.DeviceApplicationsGetter,
	informer Informer,
	resync time.Duration,
) (*ApplicationsRepository, error) {
	r := &ApplicationsRepository{
		logger: logger,
		client: client,
		cache:  make(map[string]map[string]*appsv1beta1.DeviceApplication),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new application is added
func (a *ApplicationsRepository) OnAdd(obj interface{}) {
	app, ok := obj.(*appsv1beta1.DeviceApplication)
	if !ok {
		a.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.cache[app.Namespace]; !ok {
		a.cache[app.Namespace] = make(map[string]*appsv1beta1.DeviceApplication)
	}

	a.cache[app.Namespace][app.Name] = app
}

// OnUpdate is called by the informer when an application is updated
func (a *ApplicationsRepository) OnUpdate(oldObj, newObj interface{}) {
	app, ok := newObj.(*appsv1beta1.DeviceApplication)
	if !ok {
		a.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.cache[app.Namespace][app.Name] = app
}

// OnDelete is called by the informer when an application is deleted
func (a *ApplicationsRepository) OnDelete(obj interface{}) {
	app, ok := obj.(*appsv1beta1.DeviceApplication)
	if !ok {
		a.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.cache[app.Namespace], app.Name)
}

// Create creates a new application
func (a *ApplicationsRepository) Create(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
	return a.client.DeviceApplications(application.Namespace).Create(application)
}

// Get returns the application with the given name if found
func (a *ApplicationsRepository) Get(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceApplication, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if app, ok := a.cache[namespace][name]; ok {
		return app, nil
	}

	return nil, fmt.Errorf("not found")
}

// List lists all applications in a given namespace
func (a *ApplicationsRepository) List(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceApplication, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if apps, ok := a.cache[namespace]; ok {
		var result []*appsv1beta1.DeviceApplication
		for _, a := range apps {
			result = append(result, a)
		}
		return result, nil
	}

	return []*appsv1beta1.DeviceApplication{}, nil
}

// Delete deletes the application with the given name
func (a *ApplicationsRepository) Delete(ctx context.Context, namespace, name string) error {
	return a.client.DeviceApplications(namespace).Delete(name, nil)
}
