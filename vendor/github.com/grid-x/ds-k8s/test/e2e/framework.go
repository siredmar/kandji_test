package e2e

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	"github.com/grid-x/ds-k8s/pkg/client/clientset/versioned"
)

const (
	managerPodName = "ds-k8s-manager"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
)

type cleanupFunc func(f *Framework)

type Framework struct {
	logger log.FieldLogger

	rnd *rand.Rand

	managerImage     string
	managerNamespace string

	kubeClient kubernetes.Interface
	crClient   versioned.Interface

	// mu protects internal state in case the framework is used concurrently
	// by multiple tests
	mu       sync.Mutex
	cleanups []cleanupFunc
}

func NewFramework(
	managerImage, managerNamespace string,
	kubeClient kubernetes.Interface,
	crClient versioned.Interface,
) (*Framework, error) {
	return &Framework{
		logger:           log.New().WithField("component", "framework"),
		rnd:              rand.New(rand.NewSource(time.Now().UnixNano())),
		managerImage:     managerImage,
		managerNamespace: managerNamespace,

		kubeClient: kubeClient,
		crClient:   crClient,
	}, nil
}

func (f *Framework) Setup() error {
	if f.managerImage != "" {

		// create namespace
		_, err := f.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.managerNamespace,
			},
		})
		if err != nil {
			f.logger.Warnf("was not able to create namespace: %+v", err)
		}

		cmd := []string{"/root/manager"}
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      managerPodName,
				Namespace: f.managerNamespace,
				Labels:    map[string]string{"name": "manager"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "manager",
						Image:           f.managerImage,
						ImagePullPolicy: v1.PullIfNotPresent,
						Command:         cmd,
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
			},
		}

		p, err := f.kubeClient.CoreV1().Pods(f.managerNamespace).Create(pod)
		if err != nil {
			return fmt.Errorf("error while trying to create manager pod: %+v", err)
		}
		timeout := 2 * time.Minute
		deadline := time.Now().Add(timeout)
		for p.Spec.NodeName == "" {
			p, err = f.kubeClient.CoreV1().Pods(f.managerNamespace).Get(managerPodName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error while trying to get pod: %+v", err)
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("scheduling is taking to long")
			}
		}
		f.logger.Infof("ds-k8s-manager pod is scheduled on node (%s)", p.Spec.NodeName)
		deadline = time.Now().Add(timeout)
		for p.Status.Phase != v1.PodRunning {
			p, err = f.kubeClient.CoreV1().Pods(f.managerNamespace).Get(managerPodName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error while trying to get pod: %+v", err)
			}
			if time.Now().After(deadline) {
				// dump full status
				f.logger.Infof("full pod status: %+v", p.Status)
				// dump logs
				req := f.kubeClient.CoreV1().
					Pods(f.managerNamespace).
					GetLogs(managerPodName, &v1.PodLogOptions{})

				readCloser, err := req.Stream()
				if err != nil {
					f.logger.Errorf("cannot fetch logs: %+v", err)
				} else {
					defer readCloser.Close()
					buf := bytes.NewBuffer(nil)
					_, err = io.Copy(buf, readCloser)
					f.logger.Infof("logs of pod: %s", buf.String())
				}

				return fmt.Errorf("pod is still not running")
			}
		}
		f.logger.Infof("ds-k8s-manager pod is running")
		time.Sleep(30 * time.Second)
	}
	return nil
}

func (f *Framework) Dump() error {
	p, err := f.kubeClient.CoreV1().Pods(f.managerNamespace).Get(managerPodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error while trying to get pod: %+v", err)
	}
	f.logger.Infof("full pod status: %+v", p.Status)
	// dump logs
	req := f.kubeClient.CoreV1().
		Pods(f.managerNamespace).
		GetLogs(managerPodName, &v1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		f.logger.Errorf("cannot fetch logs: %+v", err)
	} else {
		defer readCloser.Close()
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, readCloser)
		f.logger.Infof("logs of pod: %s", buf.String())
	}
	return nil
}

func (f *Framework) Teardown() error {
	if f.managerImage != "" {
		return f.kubeClient.CoreV1().Pods(f.managerNamespace).
			Delete(managerPodName, nil)
	}

	f.mu.Lock()
	// run cleanups in reverse order
	for i := len(f.cleanups) - 1; i >= 0; i-- {
		f.cleanups[i](f)
	}
	f.mu.Unlock()
	return nil
}

func (f *Framework) newCleanup(fc cleanupFunc) {
	f.mu.Lock()
	f.cleanups = append(f.cleanups, fc)
	f.mu.Unlock()
}

func (f *Framework) randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[f.rnd.Intn(len(letterRunes))]
	}
	return string(b)
}

func (f *Framework) NewTestNamespace() (string, error) {
	ns := "test-" + f.randStringRunes(10)

	f.logger.Infof("creating new namespace %s", ns)
	if _, err := f.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}); err != nil {
		return "", err
	}

	f.newCleanup(func(f *Framework) {
		f.logger.Infof("deleting namespace %s", ns)
		if err := f.kubeClient.CoreV1().Namespaces().Delete(ns, nil); err != nil {
			f.logger.Warnf("error while cleaning up: %+v", err)
		}
	})

	return ns, nil
}

func (f *Framework) CRClient() versioned.Interface {
	return f.crClient
}

func (f *Framework) KubeClient() kubernetes.Interface {
	return f.kubeClient
}

// the following are helper functions to quickly setup resources to work with

func (f *Framework) CreateNewDevices(labels map[string]string, ns string, count int) ([]*corev1beta1.Device, error) {
	devices := make([]*corev1beta1.Device, count)
	for i := 0; i < count; i++ {
		dev := &corev1beta1.Device{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("device-%d", i),
				Namespace: ns,
				Labels:    labels,
			},
			Spec:   corev1beta1.DeviceSpec{},
			Status: corev1beta1.DeviceStatus{},
		}
		dev, err := f.crClient.CoreV1beta1().Devices(ns).Create(dev)
		if err != nil {
			return nil, err
		}
		devices[i] = dev
	}
	f.newCleanup(func(f *Framework) {
		f.logger.Infof("deleting devices in namespace %s", ns)
		for _, dev := range devices {
			if err := f.crClient.CoreV1beta1().Devices(dev.Namespace).Delete(dev.Name, nil); err != nil {
				f.logger.Errorf("error while deleting devices %s/%s", dev.Namespace, dev.Name)
			}
		}
	})
	return devices, nil
}
