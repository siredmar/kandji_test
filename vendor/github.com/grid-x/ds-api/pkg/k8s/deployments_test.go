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

type deployCreateF func(client *mockDeployClient, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
type deployUpdateF func(client *mockDeployClient, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
type deployDeleteF func(client *mockDeployClient, name string, options *metav1.DeleteOptions) error
type deployGetF func(client *mockDeployClient, name string, options metav1.GetOptions) (*appsv1beta1.DeviceDeployment, error)
type deployListF func(client *mockDeployClient, opts metav1.ListOptions) (*appsv1beta1.DeviceDeploymentList, error)

type mockDeployGetter struct {
	create deployCreateF
	update deployUpdateF
	delete deployDeleteF
	get    deployGetF
	list   deployListF
}

func (m *mockDeployGetter) DeviceDeployments(namespace string) appsclient.DeviceDeploymentInterface {
	return &mockDeployClient{
		namespace: namespace,
		create:    m.create,
		update:    m.update,
		delete:    m.delete,
		get:       m.get,
		list:      m.list,
	}
}

type mockDeployClient struct {
	namespace string

	create deployCreateF
	update deployUpdateF
	delete deployDeleteF
	get    deployGetF
	list   deployListF
}

func (c *mockDeployClient) Create(deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return c.create(c, deploy)
}

func (c *mockDeployClient) Update(deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return c.update(c, deploy)
}

func (c *mockDeployClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(c, name, options)
}

func (c *mockDeployClient) Get(name string, options metav1.GetOptions) (*appsv1beta1.DeviceDeployment, error) {
	return c.get(c, name, options)
}

func (c *mockDeployClient) List(opts metav1.ListOptions) (*appsv1beta1.DeviceDeploymentList, error) {
	return c.list(c, opts)
}

func (c *mockDeployClient) UpdateStatus(*appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	panic("not needed")
}

func (c *mockDeployClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}

func (c *mockDeployClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}

func (c *mockDeployClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *appsv1beta1.DeviceDeployment, err error) {
	panic("not needed")
}

type mockDeployInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockDeployInformer) add(deploy *appsv1beta1.DeviceDeployment) {
	i.handler.OnAdd(deploy)
}

func (i *mockDeployInformer) update(old, new *appsv1beta1.DeviceDeployment) {
	i.handler.OnUpdate(old, new)
}

func (i *mockDeployInformer) delete(deploy *appsv1beta1.DeviceDeployment) {
	i.handler.OnDelete(deploy)
}

func (i *mockDeployInformer) Informer() Informer {
	return i
}

func (i *mockDeployInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_Deployments_Create(t *testing.T) {
	var got *appsv1beta1.DeviceDeployment

	want := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   appsv1beta1.DeviceDeploymentSpec{},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	repo, err := NewDeploymentsRepository(
		log.New(),
		&mockDeployGetter{
			create: func(client *mockDeployClient, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
				got = deploy
				return deploy, nil
			},
		},
		&mockDeployInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create deployments repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Create(ctx, want)
	if err != nil {
		t.Fatalf("cannot create deployment: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected deployment: %s", cmp.Diff(want, got))
	}

}

func Test_Deployments_Update(t *testing.T) {
	var got *appsv1beta1.DeviceDeployment

	want := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   appsv1beta1.DeviceDeploymentSpec{},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}

	repo, err := NewDeploymentsRepository(
		log.New(),
		&mockDeployGetter{
			update: func(client *mockDeployClient, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
				got = deploy
				return deploy, nil
			},
		},
		&mockDeployInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create deployments repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Update(ctx, want)
	if err != nil {
		t.Fatalf("cannot update deployment: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected deployments: %s", cmp.Diff(want, got))
	}

}

func Test_Deployments_Get(t *testing.T) {
	want := &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   appsv1beta1.DeviceDeploymentSpec{},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}
	informer := &mockDeployInformer{}
	repo, err := NewDeploymentsRepository(
		log.New(),
		&mockDeployGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create deployments repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get deployment: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected deployment: %s", cmp.Diff(want, got))
	}
}

type DeployByName []*appsv1beta1.DeviceDeployment

func (n DeployByName) Len() int {
	return len(n)
}

func (n DeployByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n DeployByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_Deployments_List(t *testing.T) {
	want := make([]*appsv1beta1.DeviceDeployment, 10)

	for i := 0; i < 10; i++ {
		want[i] = &appsv1beta1.DeviceDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   appsv1beta1.DeviceDeploymentSpec{},
			Status: appsv1beta1.DeviceDeploymentStatus{},
		}
	}
	informer := &mockDeployInformer{}
	repo, err := NewDeploymentsRepository(
		log.New(),
		&mockDeployGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create deployments repo: %+v", err)
	}

	for _, deploy := range want {
		informer.add(deploy)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get deployment: %+v", err)
	}

	sort.Sort(DeployByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected deployment: %s", cmp.Diff(want, got))
	}
}

func Test_Deployments_Delete(t *testing.T) {
	var gotNamespace, gotName string

	wantName := "foo"
	wantNamespace := "bar"

	repo, err := NewDeploymentsRepository(
		log.New(),
		&mockDeployGetter{
			delete: func(client *mockDeployClient, name string, options *metav1.DeleteOptions) error {
				gotNamespace = client.namespace
				gotName = name
				return nil
			},
		},
		&mockDeployInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create deployments repo: %+v", err)
	}
	ctx := context.TODO()
	err = repo.Delete(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot delete deployment: %+v", err)
	}

	if !cmp.Equal(wantName, gotName) {
		t.Errorf("unexpected name: %s", cmp.Diff(wantName, gotName))
	}
	if !cmp.Equal(wantNamespace, gotNamespace) {
		t.Errorf("unexpected namespace: %s", cmp.Diff(wantNamespace, gotNamespace))
	}
}
