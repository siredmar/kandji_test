/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package devicejob

import (
	"fmt"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	batchv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

var (
	c               client.Client
	expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "restart-job", Namespace: "gridx-de"}}

	devices = []*corev1beta1.Device{gridBox001Stable}
)

const timeout = time.Second * 5

func findPods(namespacedName types.NamespacedName) ([]*corev1beta1.DevicePod, error) {
	job := &batchv1beta1.DeviceJob{}
	if err := c.Get(context.TODO(), namespacedName, job); err != nil {
		return nil, err
	}

	pods := &corev1beta1.DevicePodList{}
	if err := c.List(context.TODO(), nil, pods); err != nil {
		return nil, err
	}

	result := []*corev1beta1.DevicePod{}
	for _, pod := range pods.Items {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		for _, ref := range pod.ObjectMeta.OwnerReferences {
			if ref.UID == job.UID {
				result = append(result, &pod)
			}
		}
	}
	return result, nil
}

func haveActivePods(g *gomega.GomegaWithT, namespacedName types.NamespacedName, number int) {
	g.Eventually(func() error {
		pods, err := findPods(namespacedName)
		if err != nil {
			return err
		}
		if len(pods) != number {
			return fmt.Errorf("pod number not as expected. Want %d, but got %d", number, len(pods))
		}
		return nil
	}, timeout).Should(gomega.Succeed())
}

func updatePodForDevice(g *gomega.GomegaWithT, namespacedName types.NamespacedName, deviceID string, update func(*corev1beta1.DevicePod) error) {
	g.Expect(func() error {
		pods, err := findPods(namespacedName)
		if err != nil {
			return err
		}
		if len(pods) < 1 {
			return fmt.Errorf("no pods found")
		}

		for _, pod := range pods {
			if pod.DeletionTimestamp != nil {
				continue
			}
			if pod.Spec.DeviceID == deviceID {
				if err := c.Get(context.TODO(), types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}, pod); err != nil {
					return err
				}
				pod.Labels = map[string]string{
					"hello": "world",
				}
				if err := update(pod); err != nil {
					return err
				}
				return c.Update(context.TODO(), pod)
			}
		}
		return fmt.Errorf("pod not found for deviceID: %s", deviceID)
	}()).NotTo(gomega.HaveOccurred())
}

func hasJobProperty(g *gomega.GomegaWithT, namespacedName types.NamespacedName, check func(*batchv1beta1.DeviceJob) error) {
	g.Eventually(func() error {
		job := &batchv1beta1.DeviceJob{}
		if err := c.Get(context.TODO(), namespacedName, job); err != nil {
			return err
		}
		return check(job)
	}, timeout).Should(gomega.Succeed())
}

func checkStartTime(job *batchv1beta1.DeviceJob) error {
	if job.Status.StartTime == nil {
		return fmt.Errorf("StartTime not set")
	}

	// job should be started in the last 10s
	if job.Status.StartTime.Time.Before(time.Now().Add(-10 * time.Second)) {
		return fmt.Errorf("StartTime is too old")
	}
	return nil
}

func checkActiveCount(i int32) func(*batchv1beta1.DeviceJob) error {
	return func(job *batchv1beta1.DeviceJob) error {
		if job.Status.Active == nil {
			return fmt.Errorf("Active count not set")
		}
		if *job.Status.Active != i {
			return fmt.Errorf("Active count not as expected. Want %d, but got: %d", i, *job.Status.Active)
		}
		return nil
	}
}

func checkSucceededCount(i int32) func(*batchv1beta1.DeviceJob) error {
	return func(job *batchv1beta1.DeviceJob) error {
		if job.Status.Succeeded == nil {
			return fmt.Errorf("Succeeded count not set")
		}
		if *job.Status.Succeeded != i {
			return fmt.Errorf("Succeeded count not as expected. Want %d, but got: %d", i, *job.Status.Active)
		}
		return nil
	}
}

func checkFailedCount(i int32) func(*batchv1beta1.DeviceJob) error {
	return func(job *batchv1beta1.DeviceJob) error {
		if job.Status.Failed == nil {
			return fmt.Errorf("Failed count not set")
		}
		if *job.Status.Failed != i {
			return fmt.Errorf("Failed count not as expected. Want %d, but got: %d", i, *job.Status.Active)
		}
		return nil
	}
}
func checkConditionComplete(job *batchv1beta1.DeviceJob) error {
	if len(job.Status.Conditions) != 1 {
		return fmt.Errorf("unexpected number of conditions")
	}
	cond := job.Status.Conditions[0]
	if cond.Type != batchv1beta1.JobComplete {
		return fmt.Errorf("unexpected condition type: %s", cond.Type)
	}
	if cond.Status != corev1beta1.ConditionTrue {
		return fmt.Errorf("unexpected condition status: %+v", cond.Status)
	}
	return nil
}

func TestReconcile_Basic(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	instance := &batchv1beta1.DeviceJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "restart-job",
			Namespace: "gridx-de",
		},
		Spec: batchv1beta1.DeviceJobSpec{
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"gridx.de/channel": "stable",
				},
			},
			Template: corev1beta1.PodConfig{
				Containers: []corev1beta1.Container{
					{
						Image:   "docker.gridx.ai/foo:bar",
						Command: []string{"foo"},
					},
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	for _, d := range devices {
		err = c.Create(context.TODO(), d)
		// The instance object may not be a valid object because it might be missing some required fields.
		// Please modify the instance object by adding required fields and then remove the following if statement.
		if apierrors.IsInvalid(err) {
			t.Fatalf("failed to create object, got an invalid object error: %v", err)
			return
		}
		g.Expect(err).NotTo(gomega.HaveOccurred())
		defer c.Delete(context.Background(), d)
	}

	t.Logf("Creating devicejob...")
	// Create the DeviceJob object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Fatalf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	t.Logf("Checking if reconcile is called...")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	t.Logf("Checking if job properties...")
	hasJobProperty(g, expectedRequest.NamespacedName, checkStartTime)
	hasJobProperty(g, expectedRequest.NamespacedName, checkActiveCount(1))
	hasJobProperty(g, expectedRequest.NamespacedName, checkSucceededCount(0))
	hasJobProperty(g, expectedRequest.NamespacedName, checkFailedCount(0))

	t.Logf("Checking if pods exist...")
	haveActivePods(g, expectedRequest.NamespacedName, 1)

	// Delete the container and expect Reconcile to be called for Job deletion
	t.Logf("Deleting pod...")
	g.Expect(func() error {
		pods, err := findPods(expectedRequest.NamespacedName)
		if err != nil {
			return err
		}
		if len(pods) < 1 {
			return fmt.Errorf("no pods found")
		}
		t.Logf("deleting pod: %s", pods[0].Name)
		return c.Delete(context.Background(), pods[0])
	}()).NotTo(gomega.HaveOccurred())
	time.Sleep(2 * time.Second)

	t.Logf("Waiting for reconcile...")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	time.Sleep(2 * time.Second)

	t.Logf("Checking job properties again...")
	hasJobProperty(g, expectedRequest.NamespacedName, checkActiveCount(1))
	hasJobProperty(g, expectedRequest.NamespacedName, checkSucceededCount(0))
	hasJobProperty(g, expectedRequest.NamespacedName, checkFailedCount(0))

	t.Logf("Checking for active pods...")
	haveActivePods(g, expectedRequest.NamespacedName, 1)
	time.Sleep(2 * time.Second)

	// Update container status
	t.Logf("Updating pod...")
	updatePodForDevice(g, expectedRequest.NamespacedName, gridBox001Stable.Name, func(pod *corev1beta1.DevicePod) error {
		pod.Status.Conditions = []corev1beta1.PodCondition{
			{
				Type:               corev1beta1.PodConditionCompleted,
				Status:             corev1beta1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
			},
		}
		return nil
	})
	t.Logf("Waiting for reconcile to happen...")
	time.Sleep(2 * time.Second)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	time.Sleep(2 * time.Second)
	t.Logf("Checking job properties again...")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	hasJobProperty(g, expectedRequest.NamespacedName, checkActiveCount(0))
	hasJobProperty(g, expectedRequest.NamespacedName, checkSucceededCount(1))
	hasJobProperty(g, expectedRequest.NamespacedName, checkFailedCount(0))
	hasJobProperty(g, expectedRequest.NamespacedName, checkConditionComplete)
}
