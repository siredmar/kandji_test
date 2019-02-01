package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	appsclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/apps/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

type appCreateF func(client *mockAppClient, app *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error)
type appDeleteF func(client *mockAppClient, name string, options *metav1.DeleteOptions) error
type appGetF func(client *mockAppClient, name string, options metav1.GetOptions) (*appsv1beta1.DeviceApplication, error)
type appListF func(client *mockAppClient, opts metav1.ListOptions) (*appsv1beta1.DeviceApplicationList, error)

type mockAppGetter struct {
	create appCreateF
	delete appDeleteF
	get    appGetF
	list   appListF
}

func (m *mockAppGetter) DeviceApplications(namespace string) appsclient.DeviceApplicationInterface {
	return &mockAppClient{
		namespace: namespace,
		create:    m.create,
		delete:    m.delete,
		get:       m.get,
		list:      m.list,
	}
}

type mockAppClient struct {
	namespace string

	create appCreateF
	delete appDeleteF
	get    appGetF
	list   appListF
}

func (c *mockAppClient) Create(app *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
	return c.create(c, app)
}

func (c *mockAppClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(c, name, options)
}

func (c *mockAppClient) Get(name string, options metav1.GetOptions) (*appsv1beta1.DeviceApplication, error) {
	return c.get(c, name, options)
}

func (c *mockAppClient) List(opts metav1.ListOptions) (*appsv1beta1.DeviceApplicationList, error) {
	return c.list(c, opts)
}

func (c *mockAppClient) Update(*appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
	panic("not needed")
}

func (c *mockAppClient) UpdateStatus(*appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
	panic("not needed")
}

func (c *mockAppClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}

func (c *mockAppClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}

func (c *mockAppClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *appsv1beta1.DeviceApplication, err error) {
	panic("not needed")
}

type mockAppInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockAppInformer) add(app *appsv1beta1.DeviceApplication) {
	i.handler.OnAdd(app)
}

func (i *mockAppInformer) update(old, new *appsv1beta1.DeviceApplication) {
	i.handler.OnUpdate(old, new)
}

func (i *mockAppInformer) delete(app *appsv1beta1.DeviceApplication) {
	i.handler.OnDelete(app)
}

func (i *mockAppInformer) Informer() Informer {
	return i
}

func (i *mockAppInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_Applications_Create(t *testing.T) {
	var got *appsv1beta1.DeviceApplication

	want := &appsv1beta1.DeviceApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   appsv1beta1.DeviceApplicationSpec{},
		Status: appsv1beta1.DeviceApplicationStatus{},
	}

	repo, err := NewApplicationsRepository(
		log.New(),
		&mockAppGetter{
			create: func(client *mockAppClient, app *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
				got = app
				return app, nil
			},
		},
		&mockAppInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create applications repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Create(ctx, want)
	if err != nil {
		t.Fatalf("cannot create application: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected application: %s", cmp.Diff(want, got))
	}

}

func Test_Applications_Get(t *testing.T) {
	want := &appsv1beta1.DeviceApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   appsv1beta1.DeviceApplicationSpec{},
		Status: appsv1beta1.DeviceApplicationStatus{},
	}
	informer := &mockAppInformer{}
	repo, err := NewApplicationsRepository(
		log.New(),
		&mockAppGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create applications repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get application: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected application: %s", cmp.Diff(want, got))
	}
}

type ByName []*appsv1beta1.DeviceApplication

func (n ByName) Len() int {
	return len(n)
}

func (n ByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n ByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_Applications_List(t *testing.T) {
	want := make([]*appsv1beta1.DeviceApplication, 10)

	for i := 0; i < 10; i++ {
		want[i] = &appsv1beta1.DeviceApplication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   appsv1beta1.DeviceApplicationSpec{},
			Status: appsv1beta1.DeviceApplicationStatus{},
		}
	}
	informer := &mockAppInformer{}
	repo, err := NewApplicationsRepository(
		log.New(),
		&mockAppGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create applications repo: %+v", err)
	}

	for _, app := range want {
		informer.add(app)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get application: %+v", err)
	}

	sort.Sort(ByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected application: %s", cmp.Diff(want, got))
	}
}

func Test_Applications_Delete(t *testing.T) {
	var gotNamespace, gotName string

	wantName := "foo"
	wantNamespace := "bar"

	repo, err := NewApplicationsRepository(
		log.New(),
		&mockAppGetter{
			delete: func(client *mockAppClient, name string, options *metav1.DeleteOptions) error {
				gotNamespace = client.namespace
				gotName = name
				return nil
			},
		},
		&mockAppInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create applications repo: %+v", err)
	}
	ctx := context.TODO()
	err = repo.Delete(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot delete application: %+v", err)
	}

	if !cmp.Equal(wantName, gotName) {
		t.Errorf("unexpected name: %s", cmp.Diff(wantName, gotName))
	}
	if !cmp.Equal(wantNamespace, gotNamespace) {
		t.Errorf("unexpected namespace: %s", cmp.Diff(wantNamespace, gotNamespace))
	}
}
