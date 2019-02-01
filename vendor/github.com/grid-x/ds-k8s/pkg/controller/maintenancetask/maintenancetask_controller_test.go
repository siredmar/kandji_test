/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package maintenancetask

import (
	"testing"
	"time"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	batchv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
	maintenancev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foobar", Namespace: "default"}}
var jobKey = types.NamespacedName{Name: "foobar-job", Namespace: "default"}

const timeout = time.Second * 30

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &maintenancev1beta1.MaintenanceTask{
		ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "default"},
		Spec: maintenancev1beta1.MaintenanceTaskSpec{
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"foo": "bar",
				},
			},
			Type: maintenancev1beta1.MaintenanceTaskTypeRestart,
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the MaintenanceTask object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	job := &batchv1beta1.DeviceJob{}
	g.Eventually(func() error { return c.Get(context.TODO(), jobKey, job) }, timeout).
		Should(gomega.Succeed())

	// Manually delete job since GC isn't enabled in the test control plane
	g.Expect(c.Delete(context.TODO(), job)).To(gomega.Succeed())

}
