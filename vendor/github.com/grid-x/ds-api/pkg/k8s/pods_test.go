package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	coreclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/core/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

type podUpdateStatusF func(client *mockPodClient, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)

type mockPodGetter struct {
	updateStatus podUpdateStatusF
}

func (m *mockPodGetter) DevicePods(namespace string) coreclient.DevicePodInterface {
	return &mockPodClient{
		namespace:    namespace,
		updateStatus: m.updateStatus,
	}
}

type mockPodClient struct {
	namespace string

	updateStatus podUpdateStatusF
}

func (c *mockPodClient) UpdateStatus(pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return c.updateStatus(c, pod)
}

func (c *mockPodClient) Create(*corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	panic("not needed")
}
func (c *mockPodClient) Update(*corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	panic("not needed")
}
func (c *mockPodClient) Delete(name string, options *metav1.DeleteOptions) error {
	panic("not needed")
}
func (c *mockPodClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}
func (c *mockPodClient) Get(name string, options metav1.GetOptions) (*corev1beta1.DevicePod, error) {
	panic("not needed")
}
func (c *mockPodClient) List(opts metav1.ListOptions) (*corev1beta1.DevicePodList, error) {
	panic("not needed")
}
func (c *mockPodClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}
func (c *mockPodClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *corev1beta1.DevicePod, err error) {
	panic("not needed")
}

type mockPodInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockPodInformer) add(pod *corev1beta1.DevicePod) {
	i.handler.OnAdd(pod)
}

func (i *mockPodInformer) update(old, new *corev1beta1.DevicePod) {
	i.handler.OnUpdate(old, new)
}

func (i *mockPodInformer) delete(pod *corev1beta1.DevicePod) {
	i.handler.OnDelete(pod)
}

func (i *mockPodInformer) Informer() Informer {
	return i
}

func (i *mockPodInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_Pods_UpdateStatus(t *testing.T) {
	var got *corev1beta1.DevicePod

	want := &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DevicePodSpec{},
		Status: corev1beta1.DevicePodStatus{},
	}

	repo, err := NewPodsRepository(
		log.New(),
		&mockPodGetter{
			updateStatus: func(client *mockPodClient, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				got = pod
				return pod, nil
			},
		},
		&mockPodInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create pods repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.UpdateStatus(ctx, want)
	if err != nil {
		t.Fatalf("cannot update status of pod: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected pod: %s", cmp.Diff(want, got))
	}

}

func Test_Pods_Get(t *testing.T) {
	want := &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DevicePodSpec{},
		Status: corev1beta1.DevicePodStatus{},
	}
	informer := &mockPodInformer{}
	repo, err := NewPodsRepository(
		log.New(),
		&mockPodGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create pods repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get pod: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected pod: %s", cmp.Diff(want, got))
	}
}

type PodByName []*corev1beta1.DevicePod

func (n PodByName) Len() int {
	return len(n)
}

func (n PodByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n PodByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_Pods_List(t *testing.T) {
	want := make([]*corev1beta1.DevicePod, 10)

	for i := 0; i < 10; i++ {
		want[i] = &corev1beta1.DevicePod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   corev1beta1.DevicePodSpec{},
			Status: corev1beta1.DevicePodStatus{},
		}
	}
	informer := &mockPodInformer{}
	repo, err := NewPodsRepository(
		log.New(),
		&mockPodGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create pods repo: %+v", err)
	}

	for _, pod := range want {
		informer.add(pod)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get pods: %+v", err)
	}

	sort.Sort(PodByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected pods: %s", cmp.Diff(want, got))
	}
}

func Test_Pods_ListByDeviceID(t *testing.T) {
	want := make([]*corev1beta1.DevicePod, 10)

	for i := 0; i < 10; i++ {
		want[i] = &corev1beta1.DevicePod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec: corev1beta1.DevicePodSpec{
				DeviceID: "qaz",
			},
			Status: corev1beta1.DevicePodStatus{},
		}
	}
	informer := &mockPodInformer{}
	repo, err := NewPodsRepository(
		log.New(),
		&mockPodGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create pods repo: %+v", err)
	}

	for _, pod := range want {
		informer.add(pod)
	}
	for i := 0; i < 10; i++ {
		informer.add(&corev1beta1.DevicePod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec: corev1beta1.DevicePodSpec{
				DeviceID: "baz",
			},
			Status: corev1beta1.DevicePodStatus{},
		})
	}

	ctx := context.TODO()
	got, err := repo.ListByDeviceID(ctx, "bar", "qaz")
	if err != nil {
		t.Fatalf("cannot get pods: %+v", err)
	}

	sort.Sort(PodByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected pods: %s", cmp.Diff(want, got))
	}
}
