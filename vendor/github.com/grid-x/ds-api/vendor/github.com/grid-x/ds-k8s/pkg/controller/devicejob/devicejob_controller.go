/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package devicejob

import (
	"context"
	"time"

	"github.com/satori/go.uuid"
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

	batchv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	deployCtrl "github.com/grid-x/ds-k8s/pkg/controller/devicedeployment"
)

// Add creates a new DeviceJob Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeviceJob{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("devicejob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to DeviceJob
	err = c.Watch(&source.Kind{Type: &batchv1beta1.DeviceJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes on controlled pods, e.g. status update from supervisor
	err = c.Watch(&source.Kind{Type: &corev1beta1.DevicePod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &batchv1beta1.DeviceJob{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeviceJob{}

// ReconcileDeviceJob reconciles a DeviceJob object
type ReconcileDeviceJob struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DeviceJob object and makes changes based on the state read
// and what is in the DeviceJob.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.gridx.ai,resources=devicejobs,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDeviceJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	ctxx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// Fetch the DeviceJob instance
	instance := &batchv1beta1.DeviceJob{}
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

	firstTime := false
	// If starttime is nil let's set it first
	if instance.Status.StartTime == nil {
		now := metav1.Now()
		instance.Status.StartTime = &now
		firstTime = true
	}

	noStatus := false
	if instance.Status.SubStatuses == nil {
		instance.Status.SubStatuses = []batchv1beta1.SubStatus{}
		noStatus = true
	}

	// Reconcile overall state
	active, succeeded, failed := 0, 0, 0
	if firstTime || noStatus {
		devices, err := r.getMatchingDevices(ctx, instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		for _, dev := range devices {
			pod, err := r.createPodWithControllerRef(ctx, dev, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			instance.Status.SubStatuses = append(instance.Status.SubStatuses, batchv1beta1.SubStatus{
				DeviceID:     dev.Name,
				PodID:        pod.Name,
				LastSyncTime: metav1.Now(),
				StatusType:   batchv1beta1.SubStatusTypePending,
			})
			active++
		}
	} else {
		for i, st := range instance.Status.SubStatuses {
			ctxx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			pod := &corev1beta1.DevicePod{}
			err := r.Get(ctxx, types.NamespacedName{Namespace: instance.Namespace, Name: st.PodID}, pod)
			if err != nil {
				if errors.IsNotFound(err) {
					dev := &corev1beta1.Device{}
					ctxx, cancel := context.WithTimeout(ctx, 10*time.Second)
					defer cancel()
					if err := r.Get(ctxx, types.NamespacedName{Namespace: instance.Namespace, Name: st.DeviceID}, dev); err != nil {
						return reconcile.Result{}, err
					}
					pod, err := r.createPodWithControllerRef(ctx, dev, instance)
					if err != nil {
						return reconcile.Result{}, err
					}
					instance.Status.SubStatuses[i] = batchv1beta1.SubStatus{
						DeviceID:     st.DeviceID,
						PodID:        pod.Name,
						LastSyncTime: metav1.Now(),
						StatusType:   batchv1beta1.SubStatusTypePending,
					}
					active++
					continue
				} else {
					return reconcile.Result{}, err
				}
			}

			st := batchv1beta1.SubStatus{
				DeviceID:     st.DeviceID,
				PodID:        pod.Name,
				LastSyncTime: metav1.Now(),
				StatusType:   batchv1beta1.SubStatusTypePending,
			}
			for _, cond := range pod.Status.Conditions {
				switch cond.Type {
				case corev1beta1.PodConditionScheduled, corev1beta1.PodConditionDownloading:
					st.StatusType = batchv1beta1.SubStatusTypePending
					active++
				case corev1beta1.PodConditionCompleted:
					st.StatusType = batchv1beta1.SubStatusTypeSucceeded
					succeeded++
				case corev1beta1.PodConditionFailed:
					st.StatusType = batchv1beta1.SubStatusTypeFailed
					failed++
				default:
					st.StatusType = batchv1beta1.SubStatusTypeRunning
					active++
				}
			}
			instance.Status.SubStatuses[i] = st

		}
	}

	if instance.Status.Active == nil {
		instance.Status.Active = new(int32)
	}
	*instance.Status.Active = int32(active)

	if instance.Status.Succeeded == nil {
		instance.Status.Succeeded = new(int32)
	}
	*instance.Status.Succeeded = int32(succeeded)

	if instance.Status.Failed == nil {
		instance.Status.Failed = new(int32)
	}
	*instance.Status.Failed = int32(failed)

	if active == 0 && (succeeded > 0 || failed > 0) {
		now := metav1.Now()
		cond := batchv1beta1.JobCondition{
			Type:          batchv1beta1.JobComplete,
			Status:        corev1beta1.ConditionTrue,
			LastProbeTime: &now,
		}
		if failed > 0 {
			cond.Type = batchv1beta1.JobFailed
		}
		instance.Status.Conditions = []batchv1beta1.JobCondition{cond}
	}

	// Update status of instance
	ctxx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err = r.Update(ctxx, instance)
	return reconcile.Result{}, err
}

func (r *ReconcileDeviceJob) createPodWithControllerRef(ctx context.Context, dev *corev1beta1.Device, job *batchv1beta1.DeviceJob) (*corev1beta1.DevicePod, error) {
	pod := &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewV4().String(),
			Namespace: job.Namespace,
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: dev.Name,
			Config:   job.Spec.Template,
		},
		Status: corev1beta1.DevicePodStatus{},
	}
	if err := controllerutil.SetControllerReference(job, pod, r.scheme); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := r.Create(ctx, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *ReconcileDeviceJob) getMatchingDevices(ctx context.Context, job *batchv1beta1.DeviceJob) ([]*corev1beta1.Device, error) {
	result := []*corev1beta1.Device{}
	devices := &corev1beta1.DeviceList{}
	ctxx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := r.List(ctxx, nil, devices); err != nil {
		return nil, err
	}

	for _, dev := range devices.Items {
		if deployCtrl.Matches(dev, job.Spec.Selector) {
			result = append(result, &dev)
		}
	}

	return result, nil
}
