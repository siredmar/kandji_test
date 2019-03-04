package devicedeployment

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
)

type enqueueDeviceDeployment struct {
	cl client.Client
}

// Create is called in response to an create event - e.g. Pod Creation.
func (e *enqueueDeviceDeployment) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: evt.Meta.GetNamespace(),
		Name:      evt.Meta.GetName(),
	}})
}

// Update is called in response to an update event -  e.g. Pod Updated.
func (e *enqueueDeviceDeployment) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: evt.MetaNew.GetNamespace(),
		Name:      evt.MetaNew.GetName(),
	}})
}

// Delete is called in response to a delete event - e.g. Pod Deleted.
func (e *enqueueDeviceDeployment) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	ctx := context.Background()
	// deploy := *evt.Object.(*appsv1beta1.DeviceDeployment)

	ds := &appsv1beta1.DeviceDeploymentList{}
	if err := e.cl.List(ctx, client.InNamespace(evt.Meta.GetNamespace()), ds); err != nil {
		fmt.Println(err)
		return
	}

	for _, deploy := range ds.Items {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: deploy.Namespace,
			Name:      deploy.Name,
		}})
	}

}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
func (e *enqueueDeviceDeployment) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: evt.Meta.GetNamespace(),
		Name:      evt.Meta.GetName(),
	}})
}
