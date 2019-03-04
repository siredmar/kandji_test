package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

var (
	appsv1beta1DeviceDeploymentIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)

	defaultTimeout = 10 * time.Second
	longTimeout    = 10 * time.Minute
)

func Test_Appsv1beta1_DeviceDeployment_Basic_CRUD(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		dev     *appsv1beta1.DeviceDeployment
		update  *appsv1beta1.DeviceDeployment
	}{
		{
			wantErr: true,
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			wantErr: true,
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   appsv1beta1.DeviceDeploymentSpec{},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
		},
		{
			wantErr: true,
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
		},
		{
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{},
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
		},
		{
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{},
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
			update: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "foo",
									Image: "bar",
								},
							},
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
		},
		{
			dev: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
							"baz": "faz",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{},
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
			update: &appsv1beta1.DeviceDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "foo",
									Image: "bar",
								},
							},
						},
					},
				},
				Status: appsv1beta1.DeviceDeploymentStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.AppsV1beta1().DeviceDeployments(ns).Create(tc.dev)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			defer func() {
				if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(tc.dev.Name, nil); err != nil {
					t.Errorf("cannot delete device deployment %s/%s: %+v", ns, tc.dev.Name, err)
				}
			}()

			got, err := client.AppsV1beta1().DeviceDeployments(ns).Get(tc.dev.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device deployment %s/%s: %+v", ns, tc.dev.Name, err)
			}

			if !cmp.Equal(tc.dev, got, appsv1beta1DeviceDeploymentIgnores) {
				t.Errorf("unexpected device deployment: %s", cmp.Diff(tc.dev, got, appsv1beta1DeviceDeploymentIgnores))
			}

			if tc.update != nil {
				tc.update.ResourceVersion = got.ResourceVersion
				_, err := client.AppsV1beta1().DeviceDeployments(ns).Update(tc.update)
				if err != nil {
					t.Fatalf("error while updating deployment: %+v", err)
				}

				got, err = client.AppsV1beta1().DeviceDeployments(ns).Get(tc.dev.Name, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("error while getting device deployment: %+v", err)
				}

				if !cmp.Equal(tc.update, got, appsv1beta1DeviceDeploymentIgnores) {
					t.Errorf("unexpected device pod: %s", cmp.Diff(tc.update, got, appsv1beta1DeviceDeploymentIgnores))
				}
			}
		})
	}

}

func Test_Appsv1beta1_DeviceDeployment_Create_Simple(t *testing.T) {
	t.Parallel()
	client := GLOBAL.CRClient()
	g := gomega.NewGomegaWithT(t)

	logger := log.New().WithField("component", "test")

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	deploy := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "foo",
							Image: "bar",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	logger.Infof("creating deployment %s/%s", ns, deploy.Name)
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Create(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	numMatchingBoxes := 200

	logger.Info("creating devices")
	_, err = GLOBAL.CreateNewDevices(map[string]string{
		"apps.gridx.de/channel": "stable",
	}, ns, numMatchingBoxes)

	logger.Infof("waiting for deployment %s/%s to appear", ns, deploy.Name)
	g.Eventually(func() error {
		_, err := client.AppsV1beta1().DeviceDeployments(ns).Get(deploy.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, defaultTimeout).Should(gomega.Succeed())
	logger.Infof("deployment %s/%s ready", ns, deploy.Name)

	logger.Info("waiting for pods to appear")
	g.Eventually(func() error {
		pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if len(pods.Items) != numMatchingBoxes {
			return fmt.Errorf("not enough pods found")
		}
		return nil
	}, longTimeout).Should(gomega.Succeed())
	logger.Info("pods ready")

	defer func() {
		logger.Infof("deleting deployment %s/%s", ns, deploy.Name)
		if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy.Name, nil); err != nil {
			t.Errorf("cannot delete device deployment %s/%s: %+v", ns, deploy.Name, err)
		}
	}()
}

func Test_Appsv1beta1_DeviceDeployment_Create_Simple_Box_First(t *testing.T) {
	t.Parallel()
	client := GLOBAL.CRClient()
	g := gomega.NewGomegaWithT(t)

	logger := log.New().WithField("component", "test")

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	numMatchingBoxes := 200

	logger.Info("creating devices")
	_, err = GLOBAL.CreateNewDevices(map[string]string{
		"apps.gridx.de/channel": "stable",
	}, ns, numMatchingBoxes)

	deploy := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "foo",
							Image: "bar",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	logger.Infof("creating deployment %s/%s", ns, deploy.Name)
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Create(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	logger.Infof("waiting for deployment %s/%s to appear", ns, deploy.Name)
	g.Eventually(func() error {
		_, err := client.AppsV1beta1().DeviceDeployments(ns).Get(deploy.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, defaultTimeout).Should(gomega.Succeed())
	logger.Infof("deployment %s/%s ready", ns, deploy.Name)

	logger.Info("waiting for pods to appear")
	g.Eventually(func() error {
		pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if len(pods.Items) != numMatchingBoxes {
			return fmt.Errorf("not enough pods found")
		}
		return nil
	}, longTimeout).Should(gomega.Succeed())
	logger.Info("pods ready")

	defer func() {
		logger.Infof("deleting deployment %s/%s", ns, deploy.Name)
		if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy.Name, nil); err != nil {
			t.Errorf("cannot delete device deployment %s/%s: %+v", ns, deploy.Name, err)
		}
	}()
}

func Test_Appsv1beta1_DeviceDeployment_Create_Update(t *testing.T) {
	t.Parallel()
	client := GLOBAL.CRClient()
	g := gomega.NewGomegaWithT(t)

	logger := log.New().WithField("component", "test")

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	deploy := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "foo",
							Image: "bar",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	logger.Infof("creating deployment %s/%s", ns, deploy.Name)
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Create(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	numMatchingBoxes := 200

	logger.Info("creating devices")
	_, err = GLOBAL.CreateNewDevices(map[string]string{
		"apps.gridx.de/channel": "stable",
	}, ns, numMatchingBoxes)

	logger.Infof("waiting for deployment %s/%s to appear", ns, deploy.Name)
	g.Eventually(func() error {
		_, err := client.AppsV1beta1().DeviceDeployments(ns).Get(deploy.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, defaultTimeout).Should(gomega.Succeed())
	logger.Infof("deployment %s/%s ready", ns, deploy.Name)

	logger.Info("waiting for pods to appear")
	g.Eventually(func() error {
		pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if len(pods.Items) != numMatchingBoxes {
			return fmt.Errorf("not enough pods found")
		}
		return nil
	}, longTimeout).Should(gomega.Succeed())
	logger.Info("pods ready")

	// change channel of deployment such that all pods get removed
	deploy.Spec.Selector.MatchByLabels["apps.gridx.de/channel"] = "alpha"
	logger.Infof("updating deployment %s/%s", ns, deploy.Name)
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Update(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	logger.Info("waiting for pods to be removed")
	g.Eventually(func() error {
		pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if len(pods.Items) != 0 {
			return fmt.Errorf("not all pods removed")
		}
		return nil
	}, longTimeout).Should(gomega.Succeed())

	defer func() {
		logger.Infof("deleting deployment %s/%s", ns, deploy.Name)
		if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy.Name, nil); err != nil {
			t.Errorf("cannot delete device deployment %s/%s: %+v", ns, deploy.Name, err)
		}
	}()
}

func Test_Appsv1beta1_DeviceDeployment_Create_UpdateImage(t *testing.T) {
	t.Parallel()
	client := GLOBAL.CRClient()
	g := gomega.NewGomegaWithT(t)

	logger := log.New().WithField("component", "test")

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	deploy := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "foo",
							Image: "bar",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	logger.Infof("creating deployment %s/%s", ns, deploy.Name)
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Create(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	numMatchingBoxes := 200

	logger.Info("creating devices")
	_, err = GLOBAL.CreateNewDevices(map[string]string{
		"apps.gridx.de/channel": "stable",
	}, ns, numMatchingBoxes)

	logger.Infof("waiting for deployment %s/%s to appear", ns, deploy.Name)
	g.Eventually(func() error {
		_, err := client.AppsV1beta1().DeviceDeployments(ns).Get(deploy.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, defaultTimeout).Should(gomega.Succeed())
	logger.Infof("deployment %s/%s ready", ns, deploy.Name)
	validate := func(deploy *appsv1beta1.DeviceDeployment, num int) func() error {
		return func() error {
			pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
			if err != nil {
				return err
			}
			if len(pods.Items) != num {
				return fmt.Errorf("not enough pods found")
			}

			for _, pod := range pods.Items {
				want := deploy.Spec.Template.Spec
				got := pod.Spec.Config
				if !cmp.Equal(want, got) {
					return fmt.Errorf("invalid pod config for pod %s/%s: %s", pod.Namespace, pod.Name, cmp.Diff(want, got))
				}
			}
			return nil
		}
	}
	logger.Info("waiting for pods to appear")
	g.Eventually(validate(deploy, numMatchingBoxes), longTimeout).Should(gomega.Succeed())
	logger.Info("pods ready")

	logger.Infof("updating deployment %s/%s", ns, deploy.Name)
	deploy.Spec.Template.Spec.Containers[0].Image = "baz"
	deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Update(deploy)
	if err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	logger.Info("waiting for pods to changed")
	g.Eventually(validate(deploy, numMatchingBoxes), longTimeout).Should(gomega.Succeed())

	defer func() {
		logger.Infof("deleting deployment %s/%s", ns, deploy.Name)
		if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy.Name, nil); err != nil {
			t.Errorf("cannot delete device deployment %s/%s: %+v", ns, deploy.Name, err)
		}
	}()
}

// This tests creates two deployments which both match. We will delete one and
// expect the other one to take over after a while
func Test_Appsv1beta1_DeviceDeployment_Delete_TakeOver(t *testing.T) {
	t.Parallel()
	client := GLOBAL.CRClient()
	g := gomega.NewGomegaWithT(t)

	logger := log.New().WithField("component", "test")

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	deploy1 := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "foo",
							Image: "bar",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}
	deploy2 := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ems-latest-stable",
			Namespace: ns,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			App: "monitoring",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"apps.gridx.de/channel": "stable",
					"features.gridx.de/ems": "true",
				},
			},
			Template: appsv1beta1.PodTemplate{
				Spec: corev1beta1.PodConfig{
					Containers: []corev1beta1.Container{
						{
							Name:  "baz",
							Image: "qaz",
						},
					},
				},
			},
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	for _, deploy := range []*appsv1beta1.DeviceDeployment{deploy1, deploy2} {
		logger.Infof("creating deployment %s/%s", ns, deploy.Name)
		deploy, err = client.AppsV1beta1().DeviceDeployments(ns).Create(deploy)
		if err != nil {
			t.Fatalf("error while creating deployment: %+v", err)
		}
	}

	numMatchingBoxes := 20

	logger.Info("creating devices")
	_, err = GLOBAL.CreateNewDevices(map[string]string{
		"apps.gridx.de/channel": "stable",
		"features.gridx.de/ems": "true",
	}, ns, numMatchingBoxes)

	for _, deploy := range []*appsv1beta1.DeviceDeployment{deploy1, deploy2} {
		logger.Infof("waiting for deployment %s/%s to appear", ns, deploy.Name)
		g.Eventually(func() error {
			_, err := client.AppsV1beta1().DeviceDeployments(ns).Get(deploy.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			return nil
		}, defaultTimeout).Should(gomega.Succeed())
		logger.Infof("deployment %s/%s ready", ns, deploy.Name)
	}

	validate := func(deploy *appsv1beta1.DeviceDeployment, num int) func() error {
		return func() error {
			pods, err := client.CoreV1beta1().DevicePods(ns).List(metav1.ListOptions{})
			if err != nil {
				return err
			}
			if len(pods.Items) != num {
				return fmt.Errorf("not enough pods found")
			}

			for _, pod := range pods.Items {
				want := deploy.Spec.Template.Spec
				got := pod.Spec.Config
				if !cmp.Equal(want, got) {
					return fmt.Errorf("invalid pod config for pod %s/%s: %s", pod.Namespace, pod.Name, cmp.Diff(want, got))
				}
			}
			return nil
		}
	}
	logger.Info("waiting for pods to appear")
	g.Eventually(validate(deploy2, numMatchingBoxes), longTimeout).Should(gomega.Succeed())
	logger.Info("pods ready")

	logger.Infof("deleting deployment %s/%s", ns, deploy2.Name)
	if err = client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy2.Name, nil); err != nil {
		t.Fatalf("error while creating deployment: %+v", err)
	}

	logger.Info("waiting for pods to changed")
	g.Eventually(validate(deploy1, numMatchingBoxes), longTimeout).Should(gomega.Succeed())

	defer func() {
		logger.Infof("deleting deployment %s/%s", ns, deploy1.Name)
		if err := client.AppsV1beta1().DeviceDeployments(ns).Delete(deploy1.Name, nil); err != nil {
			t.Errorf("cannot delete device deployment %s/%s: %+v", ns, deploy1.Name, err)
		}
	}()
}
