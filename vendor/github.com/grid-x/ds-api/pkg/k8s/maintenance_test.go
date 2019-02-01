package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	mainclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/maintenance/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

type taskCreateF func(client *mockTaskClient, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error)
type taskDeleteF func(client *mockTaskClient, name string, options *metav1.DeleteOptions) error
type taskGetF func(client *mockTaskClient, name string, options metav1.GetOptions) (*mainv1beta1.MaintenanceTask, error)
type taskListF func(client *mockTaskClient, opts metav1.ListOptions) (*mainv1beta1.MaintenanceTaskList, error)

type mockTaskGetter struct {
	create taskCreateF
	delete taskDeleteF
	get    taskGetF
	list   taskListF
}

func (m *mockTaskGetter) MaintenanceTasks(namespace string) mainclient.MaintenanceTaskInterface {
	return &mockTaskClient{
		namespace: namespace,
		create:    m.create,
		delete:    m.delete,
		get:       m.get,
		list:      m.list,
	}
}

type mockTaskClient struct {
	namespace string

	create taskCreateF
	delete taskDeleteF
	get    taskGetF
	list   taskListF
}

func (c *mockTaskClient) Create(task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	return c.create(c, task)
}

func (c *mockTaskClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(c, name, options)
}

func (c *mockTaskClient) Get(name string, options metav1.GetOptions) (*mainv1beta1.MaintenanceTask, error) {
	return c.get(c, name, options)
}

func (c *mockTaskClient) List(opts metav1.ListOptions) (*mainv1beta1.MaintenanceTaskList, error) {
	return c.list(c, opts)
}

func (c *mockTaskClient) Update(*mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	panic("not needed")
}

func (c *mockTaskClient) UpdateStatus(*mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	panic("not needed")
}

func (c *mockTaskClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}

func (c *mockTaskClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}

func (c *mockTaskClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *mainv1beta1.MaintenanceTask, err error) {
	panic("not needed")
}

type mockTaskInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockTaskInformer) add(task *mainv1beta1.MaintenanceTask) {
	i.handler.OnAdd(task)
}

func (i *mockTaskInformer) update(old, new *mainv1beta1.MaintenanceTask) {
	i.handler.OnUpdate(old, new)
}

func (i *mockTaskInformer) delete(task *mainv1beta1.MaintenanceTask) {
	i.handler.OnDelete(task)
}

func (i *mockTaskInformer) Informer() Informer {
	return i
}

func (i *mockTaskInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_MaintenanceTasks_Create(t *testing.T) {
	var got *mainv1beta1.MaintenanceTask

	want := &mainv1beta1.MaintenanceTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   mainv1beta1.MaintenanceTaskSpec{},
		Status: mainv1beta1.MaintenanceTaskStatus{},
	}

	repo, err := NewMaintenanceTasksRepository(
		log.New(),
		&mockTaskGetter{
			create: func(client *mockTaskClient, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
				got = task
				return task, nil
			},
		},
		&mockTaskInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create maintenanceTasks repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Create(ctx, want)
	if err != nil {
		t.Fatalf("cannot create maintenanceTask: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected maintenanceTask: %s", cmp.Diff(want, got))
	}

}

func Test_MaintenanceTasks_Get(t *testing.T) {
	want := &mainv1beta1.MaintenanceTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   mainv1beta1.MaintenanceTaskSpec{},
		Status: mainv1beta1.MaintenanceTaskStatus{},
	}
	informer := &mockTaskInformer{}
	repo, err := NewMaintenanceTasksRepository(
		log.New(),
		&mockTaskGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create maintenanceTasks repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get maintenanceTask: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected maintenanceTask: %s", cmp.Diff(want, got))
	}
}

type TaskByName []*mainv1beta1.MaintenanceTask

func (n TaskByName) Len() int {
	return len(n)
}

func (n TaskByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n TaskByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_MaintenanceTasks_List(t *testing.T) {
	want := make([]*mainv1beta1.MaintenanceTask, 10)

	for i := 0; i < 10; i++ {
		want[i] = &mainv1beta1.MaintenanceTask{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   mainv1beta1.MaintenanceTaskSpec{},
			Status: mainv1beta1.MaintenanceTaskStatus{},
		}
	}
	informer := &mockTaskInformer{}
	repo, err := NewMaintenanceTasksRepository(
		log.New(),
		&mockTaskGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create maintenanceTasks repo: %+v", err)
	}

	for _, task := range want {
		informer.add(task)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get maintenanceTask: %+v", err)
	}

	sort.Sort(TaskByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected maintenanceTask: %s", cmp.Diff(want, got))
	}
}

func Test_MaintenanceTasks_Delete(t *testing.T) {
	var gotNamespace, gotName string

	wantName := "foo"
	wantNamespace := "bar"

	repo, err := NewMaintenanceTasksRepository(
		log.New(),
		&mockTaskGetter{
			delete: func(client *mockTaskClient, name string, options *metav1.DeleteOptions) error {
				gotNamespace = client.namespace
				gotName = name
				return nil
			},
		},
		&mockTaskInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create maintenanceTasks repo: %+v", err)
	}
	ctx := context.TODO()
	err = repo.Delete(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot delete maintenanceTask: %+v", err)
	}

	if !cmp.Equal(wantName, gotName) {
		t.Errorf("unexpected name: %s", cmp.Diff(wantName, gotName))
	}
	if !cmp.Equal(wantNamespace, gotNamespace) {
		t.Errorf("unexpected namespace: %s", cmp.Diff(wantNamespace, gotNamespace))
	}
}
