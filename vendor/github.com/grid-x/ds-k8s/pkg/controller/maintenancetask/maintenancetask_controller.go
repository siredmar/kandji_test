/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package maintenancetask

import (
	"context"
	"fmt"
	"log"
	"time"

	batchv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	maintenancev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// TODO: update this
	defaultImage   = "gridx.de/foobar"
	defaultTimeout = 10 * time.Second
)

// Add creates a new MaintenanceTask Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMaintenanceTask{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("maintenancetask-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to MaintenanceTask
	err = c.Watch(&source.Kind{Type: &maintenancev1beta1.MaintenanceTask{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &batchv1beta1.DeviceJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &maintenancev1beta1.MaintenanceTask{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMaintenanceTask{}

// ReconcileMaintenanceTask reconciles a MaintenanceTask object
type ReconcileMaintenanceTask struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MaintenanceTask object and makes changes based on the state read
// and what is in the MaintenanceTask.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=batch.gridx.ai,resources=devicejobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maintenance.gridx.ai,resources=maintenancetasks,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileMaintenanceTask) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the MaintenanceTask instance

	ctx := context.Background()
	ctxx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	instance := &maintenancev1beta1.MaintenanceTask{}
	err := r.Get(ctxx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	job := &batchv1beta1.DeviceJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-job",
			Namespace: instance.Namespace,
		},
		Spec: batchv1beta1.DeviceJobSpec{
			Selector: instance.Spec.Selector,
			Template: corev1beta1.PodConfig{
				Containers: []corev1beta1.Container{
					{
						Name:  instance.Name + "-pod",
						Image: defaultImage,
					},
				},
				RestartPolicy: &corev1beta1.RestartPolicy{
					Type: corev1beta1.RestartPolicyOnFailure,
				},
			},
		},
	}

	switch t := instance.Spec.Type; t {
	case maintenancev1beta1.MaintenanceTaskTypeRestart:
		job.Spec.Template.Containers[0].Args = []string{"restart"}
	case maintenancev1beta1.MaintenanceTaskTypeShutdown:
		job.Spec.Template.Containers[0].Args = []string{"shutdown"}
	default:
		return reconcile.Result{}, fmt.Errorf("unknown maintenance task type: %s", t)
	}

	if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &batchv1beta1.DeviceJob{}
	ctxx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	err = r.Get(ctxx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("creating job %s/%s: %+v\n", job.Namespace, job.Name, job)
		ctxx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
		err = r.Create(ctxx, job)
		if err != nil {
			return reconcile.Result{}, err
		}
		found = job
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if instance.Status.StartedAt != nil {
		t := new(metav1.Time)
		*t = metav1.Now()
		instance.Status.StartedAt = t
	}

	// Aggregate status
	if found.Status.Active != nil {
		instance.Status.Running = int(*found.Status.Active)
	}
	if found.Status.Succeeded != nil {
		instance.Status.Successful = int(*found.Status.Succeeded)
	}
	if found.Status.Failed != nil {
		instance.Status.Failed = int(*found.Status.Failed)
	}

	ctxx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	if err = r.Update(ctxx, job); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
