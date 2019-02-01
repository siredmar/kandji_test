package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	mainclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/maintenance/v1beta1"
	log "github.com/sirupsen/logrus"
)

// MaintenanceTasksRepository provides methods to manage maintenanceTasks
type MaintenanceTasksRepository struct {
	logger log.FieldLogger
	client mainclient.MaintenanceTasksGetter

	// cache holds a mapping from namespace to map from name to maintenanceTask
	// it is protected by the mutex
	mutex sync.RWMutex
	cache map[string]map[string]*mainv1beta1.MaintenanceTask
}

// NewMaintenanceTasksRepository creates a new maintenanceTasks repository with the given
// settings
func NewMaintenanceTasksRepository(
	logger log.FieldLogger,
	client mainclient.MaintenanceTasksGetter,
	informer Informer,
	resync time.Duration,
) (*MaintenanceTasksRepository, error) {
	r := &MaintenanceTasksRepository{
		logger: logger,
		client: client,
		cache:  make(map[string]map[string]*mainv1beta1.MaintenanceTask),
	}

	informer.AddEventHandlerWithResyncPeriod(r, resync)

	return r, nil
}

// Event handlers required to implement the ResourceHandler interface

// OnAdd is called by the informer when a new maintenanceTask is added
func (d *MaintenanceTasksRepository) OnAdd(obj interface{}) {
	task, ok := obj.(*mainv1beta1.MaintenanceTask)
	if !ok {
		d.logger.Warnf("cannot not convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.cache[task.Namespace]; !ok {
		d.cache[task.Namespace] = make(map[string]*mainv1beta1.MaintenanceTask)
	}

	d.cache[task.Namespace][task.Name] = task
}

// OnUpdate is called by the informer when an existing maintenanceTask is updated
func (d *MaintenanceTasksRepository) OnUpdate(oldObj, newObj interface{}) {
	task, ok := newObj.(*mainv1beta1.MaintenanceTask)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", newObj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[task.Namespace][task.Name] = task
}

// OnDelete is called by the informer when a maintenanceTask is deleted
func (d *MaintenanceTasksRepository) OnDelete(obj interface{}) {
	task, ok := obj.(*mainv1beta1.MaintenanceTask)
	if !ok {
		d.logger.Warnf("cannot convert: %+v", obj)
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.cache[task.Namespace], task.Name)
}

// Create creates a new maintenanceTask
func (d *MaintenanceTasksRepository) Create(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	return d.client.MaintenanceTasks(task.Namespace).Create(task)
}

// Get gets the maintenanceTask with the given name in the given namespace
func (d *MaintenanceTasksRepository) Get(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if task, ok := d.cache[namespace][name]; ok {
		return task, nil
	}

	return nil, fmt.Errorf("not found")
}

// List returns the list of maintenanceTasks in the given namespace
func (d *MaintenanceTasksRepository) List(ctx context.Context, namespace string) ([]*mainv1beta1.MaintenanceTask, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if tasks, ok := d.cache[namespace]; ok {
		var result []*mainv1beta1.MaintenanceTask
		for _, task := range tasks {
			result = append(result, task)
		}
		return result, nil
	}

	return []*mainv1beta1.MaintenanceTask{}, nil
}

// Delete deletes the maintenanceTask with the given name in the given namespace
func (d *MaintenanceTasksRepository) Delete(ctx context.Context, namespace, name string) error {
	return d.client.MaintenanceTasks(namespace).Delete(name, nil)
}
