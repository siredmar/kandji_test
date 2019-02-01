package devicedeployment

import (
	"fmt"
	"testing"
	"time"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "monitoring-stable", Namespace: "gridx-de"}}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	device := gridBox001Stable
	instance := monitoringStableDeployment

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	for _, d := range []*corev1beta1.Device{device} {
		// Create the DeviceDeployment object and expect the Reconcile and Deployment to be created
		err = c.Create(context.Background(), d)
		// The instance object may not be a valid object because it might be missing some required fields.
		// Please modify the instance object by adding required fields and then remove the following if statement.
		if apierrors.IsInvalid(err) {
			t.Fatalf("failed to create object, got an invalid object error: %v", err)
			return
		}
		g.Expect(err).NotTo(gomega.HaveOccurred())
		defer c.Delete(context.Background(), d)
	}

	// Create the DeviceDeployment object and expect the Reconcile and Deployment to be created
	err = c.Create(context.Background(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Fatalf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.Background(), instance)

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	var findContainer = func() (*corev1beta1.DevicePod, error) {
		deploy := &appsv1beta1.DeviceDeployment{}
		if err := c.Get(context.Background(), expectedRequest.NamespacedName, deploy); err != nil {
			return nil, err
		}

		containers := &corev1beta1.DevicePodList{}
		if err := c.List(context.Background(), nil, containers); err != nil {
			return nil, err
		}

		for _, c := range containers.Items {
			for _, ref := range c.ObjectMeta.OwnerReferences {
				if ref.UID == deploy.UID {
					return &c, nil
				}
			}
		}
		return nil, fmt.Errorf("not found yet")

	}
	g.Eventually(func() error {
		_, err := findContainer()
		return err
	}, timeout).Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	g.Expect(func() error {
		container, err := findContainer()
		if err != nil {
			return err
		}
		return c.Delete(context.Background(), container)
	}()).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error {
		_, err := findContainer()
		return err
	}, timeout).Should(gomega.Succeed())
}
