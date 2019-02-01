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

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
}

type deviceCreateF func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error)
type deviceUpdateF func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error)
type deviceUpdateStatusF func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error)
type deviceDeleteF func(client *mockDeviceClient, name string, options *metav1.DeleteOptions) error
type deviceGetF func(client *mockDeviceClient, name string, options metav1.GetOptions) (*corev1beta1.Device, error)
type deviceListF func(client *mockDeviceClient, opts metav1.ListOptions) (*corev1beta1.DeviceList, error)

type mockDeviceGetter struct {
	create       deviceCreateF
	update       deviceUpdateF
	updateStatus deviceUpdateStatusF
	delete       deviceDeleteF
	get          deviceGetF
	list         deviceListF
}

func (m *mockDeviceGetter) Devices(namespace string) coreclient.DeviceInterface {
	return &mockDeviceClient{
		namespace:    namespace,
		create:       m.create,
		update:       m.update,
		updateStatus: m.updateStatus,
		delete:       m.delete,
		get:          m.get,
		list:         m.list,
	}
}

type mockDeviceClient struct {
	namespace string

	create       deviceCreateF
	update       deviceUpdateF
	updateStatus deviceUpdateStatusF
	delete       deviceDeleteF
	get          deviceGetF
	list         deviceListF
}

func (c *mockDeviceClient) Create(device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return c.create(c, device)
}

func (c *mockDeviceClient) Update(device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return c.update(c, device)
}

func (c *mockDeviceClient) UpdateStatus(device *corev1beta1.Device) (*corev1beta1.Device, error) {
	return c.updateStatus(c, device)
}

func (c *mockDeviceClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(c, name, options)
}

func (c *mockDeviceClient) Get(name string, options metav1.GetOptions) (*corev1beta1.Device, error) {
	return c.get(c, name, options)
}

func (c *mockDeviceClient) List(opts metav1.ListOptions) (*corev1beta1.DeviceList, error) {
	return c.list(c, opts)
}

func (c *mockDeviceClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}

func (c *mockDeviceClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}
func (c *mockDeviceClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *corev1beta1.Device, err error) {
	panic("not needed")
}

type mockDeviceInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockDeviceInformer) add(device *corev1beta1.Device) {
	i.handler.OnAdd(device)
}

func (i *mockDeviceInformer) update(old, new *corev1beta1.Device) {
	i.handler.OnUpdate(old, new)
}

func (i *mockDeviceInformer) delete(device *corev1beta1.Device) {
	i.handler.OnDelete(device)
}

func (i *mockDeviceInformer) Informer() Informer {
	return i
}

func (i *mockDeviceInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_Devices_Create(t *testing.T) {
	var got *corev1beta1.Device

	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}

	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{
			create: func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error) {
				got = device
				return device, nil
			},
		},
		&mockDeviceInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Create(ctx, want)
	if err != nil {
		t.Fatalf("cannot create Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Device: %s", cmp.Diff(want, got))
	}

}

func Test_Devices_Update(t *testing.T) {
	var got *corev1beta1.Device

	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}

	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{
			update: func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error) {
				got = device
				return device, nil
			},
		},
		&mockDeviceInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Update(ctx, want)
	if err != nil {
		t.Fatalf("cannot update Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Devices: %s", cmp.Diff(want, got))
	}

}

func Test_Devices_UpdateStatus(t *testing.T) {
	var got *corev1beta1.Device

	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}

	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{
			updateStatus: func(client *mockDeviceClient, device *corev1beta1.Device) (*corev1beta1.Device, error) {
				got = device
				return device, nil
			},
		},
		&mockDeviceInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.UpdateStatus(ctx, want)
	if err != nil {
		t.Fatalf("cannot update Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Devices: %s", cmp.Diff(want, got))
	}
}

func Test_Devices_Get(t *testing.T) {
	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{},
	}
	informer := &mockDeviceInformer{}
	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Device: %s", cmp.Diff(want, got))
	}
}

func Test_Devices_GetByPublicKey(t *testing.T) {
	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: corev1beta1.DeviceSpec{
			PublicKey: mkString("foobar"),
		},
		Status: corev1beta1.DeviceStatus{},
	}
	informer := &mockDeviceInformer{}
	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.GetByPublicKey(ctx, "foobar")
	if err != nil {
		t.Fatalf("cannot get Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Device: %s", cmp.Diff(want, got))
	}
}

func Test_Devices_GetBySerialnumber(t *testing.T) {
	unwanted := make([]*corev1beta1.Device, 10)

	for i := 0; i < 10; i++ {
		unwanted[i] = &corev1beta1.Device{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec: corev1beta1.DeviceSpec{
				Serialnumber: fmt.Sprintf("123456789-%02d", i),
			},
			Status: corev1beta1.DeviceStatus{},
		}
	}

	want := &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: corev1beta1.DeviceSpec{
			Serialnumber: "123456789",
		},
		Status: corev1beta1.DeviceStatus{},
	}
	informer := &mockDeviceInformer{}
	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	for _, device := range unwanted {
		informer.add(device)
	}
	informer.add(want)

	ctx := context.TODO()
	got, err := repo.GetBySerialnumber(ctx, "123456789")
	if err != nil {
		t.Fatalf("cannot get Device: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Device: %s", cmp.Diff(want, got))
	}
}

type DeviceByName []*corev1beta1.Device

func (n DeviceByName) Len() int {
	return len(n)
}

func (n DeviceByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n DeviceByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_Devices_List(t *testing.T) {
	want := make([]*corev1beta1.Device, 10)

	for i := 0; i < 10; i++ {
		want[i] = &corev1beta1.Device{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   corev1beta1.DeviceSpec{},
			Status: corev1beta1.DeviceStatus{},
		}
	}
	informer := &mockDeviceInformer{}
	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}

	for _, device := range want {
		informer.add(device)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get Device: %+v", err)
	}

	sort.Sort(DeviceByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected Device: %s", cmp.Diff(want, got))
	}
}

func Test_Devices_Delete(t *testing.T) {
	var gotNamespace, gotName string

	wantName := "foo"
	wantNamespace := "bar"

	repo, err := NewDevicesRepository(
		log.New(),
		&mockDeviceGetter{
			delete: func(client *mockDeviceClient, name string, options *metav1.DeleteOptions) error {
				gotNamespace = client.namespace
				gotName = name
				return nil
			},
		},
		&mockDeviceInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create Devices repo: %+v", err)
	}
	ctx := context.TODO()
	err = repo.Delete(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot delete Device: %+v", err)
	}

	if !cmp.Equal(wantName, gotName) {
		t.Errorf("unexpected name: %s", cmp.Diff(wantName, gotName))
	}
	if !cmp.Equal(wantNamespace, gotNamespace) {
		t.Errorf("unexpected namespace: %s", cmp.Diff(wantNamespace, gotNamespace))
	}
}
