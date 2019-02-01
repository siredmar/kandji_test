package devicedeployment

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

var (
	controllerKind = schema.GroupVersionKind{Group: "apps.gridx.ai", Version: "v1beta1", Kind: "DeviceDeployment"}
)

// Add creates a new DeviceDeployment Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this apps.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeviceDeployment{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func findMatchingDeploymentsWithClient(ctx context.Context, client client.Client, dev corev1beta1.Device) (map[string]appsv1beta1.DeviceDeployment, error) {
	ds := &appsv1beta1.DeviceDeploymentList{}
	if err := client.List(ctx, nil, ds); err != nil {
		return nil, err
	}
	return findMatchingDeployments(dev, ds.Items), nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("devicedeployment-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              r,
	})
	if err != nil {
		return err
	}

	// Watch for changes to DeviceDeployment
	err = c.Watch(&source.Kind{Type: &appsv1beta1.DeviceDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Enqueue reconciliation for owner of device container
	err = c.Watch(&source.Kind{Type: &corev1beta1.DevicePod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1beta1.DeviceDeployment{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{
		Type: &corev1beta1.Device{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {

				reqs := []reconcile.Request{}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				ds, err := findMatchingDeploymentsWithClient(ctx, mgr.GetClient(), *a.Object.(*corev1beta1.Device))
				if err != nil {
					log.Printf("ERROR: %+v", err)
					return reqs
				}
				for _, v := range ds {
					reqs = append(reqs, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      v.Name,
							Namespace: v.Namespace,
						},
					})
				}
				return reqs

			}),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeviceDeployment{}

// ReconcileDeviceDeployment reconciles a DeviceDeployment object
type ReconcileDeviceDeployment struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DeviceDeployment object and makes changes based on the state read
// and what is in the DeviceDeployment.Spec
// NOTE: Currently, only one instance of syncHandler can be run in parallel as
// it will also update resources owned by other deployments than the one
// identified by key. This is necessary to due the shadowing functionality of
// the selectors and the constraint that only one instance of each app should
// run on the gridBox in parallel.
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gridx.ai,resources=devicedeployments,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDeviceDeployment) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Print("Starting to sync deployment")

	deploy := &appsv1beta1.DeviceDeployment{}
	err := r.Get(context.Background(), req.NamespacedName, deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Deployment %+v not found. Probably already deleted.", req.NamespacedName)
			// let GC handle it
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, err
	}

	appName := deploy.Spec.App

	// Start syncing
	// 1. List all owned containers, i.e. current state
	// 2. List of desired state
	// 3. Sync state:
	//    - Delete obsolete containers
	//    - Create new containers, but remove conflicting ones

	// 1. Try to get the current state
	// currentState contains all owned containers indexed by deviceID
	// otherCs contains all other containers with the same appName indexed
	// by deviceID
	log.Printf("Getting current state...")
	currentState, otherCs, err := r.getCurrentState(req.NamespacedName, deploy)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}
	log.Printf("current state is: %+v", currentState)

	// 2. Get desired state
	// Get list of all gridBoxes
	log.Printf("Getting desired state...")
	desiredState, err := r.getDesiredState(req.NamespacedName, deploy)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}
	log.Printf("desired state is: %+v", desiredState)

	// Sync the state
	// First delete all containers that are not wanted
	for k, cs := range currentState {
		des, ok := desiredState[k]
		alreadyExists := false
		for _, c := range cs {
			// If !ok is true there is no container in the desired
			// state so we need to delete it
			if !ok {
				log.Printf("Deleting container %s because it is not desired", c.Name)
				if err := r.Delete(context.Background(), c); err != nil {
					return reconcile.Result{Requeue: true}, err
				}

				continue
			}
			specDiffers := !reflect.DeepEqual(c.Spec.Config, des.Spec.Config)
			appDiffers := appsv1beta1.ExtractAppName(*c) != appName
			// If something has changed we need to delete the old
			// container as well
			if appDiffers || specDiffers || alreadyExists {
				log.Printf("Spec or app differs for container for %s. Deleting...", k)
				if err := r.Delete(context.Background(), c); err != nil {
					return reconcile.Result{Requeue: true}, err
				}
			} else {
				//Container exists already remove from
				// desiredState to avoid recreation later on
				delete(desiredState, k)
				alreadyExists = true
			}
		}
	}

	// Second: process whats left from the desired state, i.e. create new
	// containers, and if necessary remove conflicting containers
	for k, newC := range desiredState {
		cs, ok := otherCs[k]
		if ok {
			for _, c := range cs {
				log.Printf("Deleting container %s because it conflicts with %s", c.Name, newC.Name)
				if err := r.Delete(context.Background(), c); err != nil {
					return reconcile.Result{Requeue: true}, err
				}
			}
		}

		log.Printf("Creating new container %s", newC.Name)
		if err := r.Create(context.Background(), newC); err != nil {
			return reconcile.Result{Requeue: true}, err
		}
	}

	log.Printf("Syncing done")
	return reconcile.Result{}, nil
}

func (r *ReconcileDeviceDeployment) getCurrentState(nsname types.NamespacedName, deploy *appsv1beta1.DeviceDeployment) (map[string][]*corev1beta1.DevicePod, map[string][]*corev1beta1.DevicePod, error) {
	appName := deploy.Spec.App

	currentState := make(map[string][]*corev1beta1.DevicePod)
	allCs, controlledCs, err := r.findControlledContainers(deploy)
	log.Printf("allCs: %+v", allCs)
	log.Printf("controlledCs: %+v", controlledCs)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range controlledCs {
		id := c.Spec.DeviceID
		if id == "" {
			log.Printf("WARN: GridBoxContainer %s has no DeviceID set", c.Name)
			if err := r.Delete(context.Background(), c); err != nil {
				log.Printf("ERROR: Couldn't delete invalid GridBoxContainer %s: %+v", c.Name, err)
			}

		}
		currentState[id] = append(currentState[id], c)
	}

	// whole state represents all containers indexed by deviceID that have
	// the same appname as the current deployment
	otherCs := make(map[string][]*corev1beta1.DevicePod)
	for _, c := range allCs {
		id := c.Spec.DeviceID
		app := appsv1beta1.ExtractAppName(*c)

		if id == "" || app == "" {
			// TODO: how to handle??
			continue
		}

		cRef := metav1.GetControllerOf(c)
		if app == appName && deploy.UID != cRef.UID {
			otherCs[id] = append(otherCs[id], c)
		}
	}

	return currentState, otherCs, nil
}

func (r *ReconcileDeviceDeployment) getDesiredState(nsname types.NamespacedName, deploy *appsv1beta1.DeviceDeployment) (map[string]*corev1beta1.DevicePod, error) {
	appName := deploy.Spec.App

	devices := &corev1beta1.DeviceList{}
	if err := r.List(context.Background(), nil, devices); err != nil {
		return nil, err
	}

	desiredState := make(map[string]*corev1beta1.DevicePod)
	for _, device := range devices.Items {
		ds, err := r.findMatchingDeployments(&device)
		if err != nil {
			return nil, err
		}
		d, ok := ds[appName]
		if ok && d.UID == deploy.UID {
			desiredState[device.Name] = r.mkContainerFromDeploy(deploy, &device)
		}
	}

	return desiredState, nil
}

func (r *ReconcileDeviceDeployment) mkContainerFromDeploy(deploy *appsv1beta1.DeviceDeployment, device *corev1beta1.Device) *corev1beta1.DevicePod {
	tmp := new(bool)
	*tmp = true
	apiVersion, kind := controllerKind.ToAPIVersionAndKind()
	appName := deploy.Spec.App
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewV4().String(),
			Namespace: deploy.Namespace,
			Annotations: map[string]string{
				appsv1beta1.AppNameAnnotation: appName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiVersion,
					Kind:       kind,
					Controller: tmp,
					Name:       deploy.Name,
					UID:        deploy.UID,
				},
			},
		},
		Spec: corev1beta1.DevicePodSpec{
			DeviceID: device.Name,
			Config:   deploy.Spec.Template.Spec,
		},
		Status: corev1beta1.DevicePodStatus{},
	}
}

func (r *ReconcileDeviceDeployment) findControlledContainers(deploy *appsv1beta1.DeviceDeployment) ([]*corev1beta1.DevicePod, []*corev1beta1.DevicePod, error) {
	cs := &corev1beta1.DevicePodList{}
	if err := r.List(context.Background(), nil, cs); err != nil {
		return nil, nil, err
	}

	var result []*corev1beta1.DevicePod
	var all []*corev1beta1.DevicePod
	for _, c := range cs.Items {
		all = append(all, &c)
		controllerRef := metav1.GetControllerOf(&c)
		if controllerRef == nil {
			continue
		}
		ownerDeploy := r.resolveControllerRef(c.Namespace, controllerRef)
		if ownerDeploy == nil {
			continue
		}
		if ownerDeploy.UID == deploy.UID {
			result = append(result, &c)
		}
	}

	return all, result, nil
}

func (r *ReconcileDeviceDeployment) findMatchingDeployments(device *corev1beta1.Device) (map[string]appsv1beta1.DeviceDeployment, error) {
	ds := &appsv1beta1.DeviceDeploymentList{}
	if err := r.List(context.Background(), nil, ds); err != nil {
		return nil, err
	}
	return findMatchingDeployments(*device, ds.Items), nil
}

func (r *ReconcileDeviceDeployment) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *appsv1beta1.DeviceDeployment {
	if controllerRef == nil {
		return nil
	} else if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	d := &appsv1beta1.DeviceDeployment{}
	if err := r.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, d); err != nil {
		return nil
	} else if d == nil {
		return nil
	}

	if d.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return d
}
