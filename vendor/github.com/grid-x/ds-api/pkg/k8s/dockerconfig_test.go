package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
	configclient "github.com/grid-x/ds-k8s/pkg/client/clientset/versioned/typed/config/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

type dockerConfigCreateF func(client *mockDockerConfigClient, c *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error)
type dockerConfigDeleteF func(client *mockDockerConfigClient, name string, options *metav1.DeleteOptions) error
type dockerConfigGetF func(client *mockDockerConfigClient, name string, options metav1.GetOptions) (*configv1beta1.DockerConfig, error)
type dockerConfigListF func(client *mockDockerConfigClient, opts metav1.ListOptions) (*configv1beta1.DockerConfigList, error)

type mockDockerConfigGetter struct {
	create dockerConfigCreateF
	delete dockerConfigDeleteF
	get    dockerConfigGetF
	list   dockerConfigListF
}

func (m *mockDockerConfigGetter) DockerConfigs(namespace string) configclient.DockerConfigInterface {
	return &mockDockerConfigClient{
		namespace: namespace,
		create:    m.create,
		delete:    m.delete,
		get:       m.get,
		list:      m.list,
	}
}

type mockDockerConfigClient struct {
	namespace string

	create dockerConfigCreateF
	delete dockerConfigDeleteF
	get    dockerConfigGetF
	list   dockerConfigListF
}

func (c *mockDockerConfigClient) Create(co *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
	return c.create(c, co)
}

func (c *mockDockerConfigClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(c, name, options)
}

func (c *mockDockerConfigClient) Get(name string, options metav1.GetOptions) (*configv1beta1.DockerConfig, error) {
	return c.get(c, name, options)
}

func (c *mockDockerConfigClient) List(opts metav1.ListOptions) (*configv1beta1.DockerConfigList, error) {
	return c.list(c, opts)
}

func (c *mockDockerConfigClient) Update(*configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
	panic("not needed")
}

func (c *mockDockerConfigClient) UpdateStatus(*configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
	panic("not needed")
}

func (c *mockDockerConfigClient) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	panic("not needed")
}

func (c *mockDockerConfigClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	panic("not needed")
}

func (c *mockDockerConfigClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *configv1beta1.DockerConfig, err error) {
	panic("not needed")
}

type mockDockerConfigInformer struct {
	handler cache.ResourceEventHandler
}

func (i *mockDockerConfigInformer) add(c *configv1beta1.DockerConfig) {
	i.handler.OnAdd(c)
}

func (i *mockDockerConfigInformer) update(old, new *configv1beta1.DockerConfig) {
	i.handler.OnUpdate(old, new)
}

func (i *mockDockerConfigInformer) delete(c *configv1beta1.DockerConfig) {
	i.handler.OnDelete(c)
}

func (i *mockDockerConfigInformer) Informer() Informer {
	return i
}

func (i *mockDockerConfigInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) {
	i.handler = handler
}

func Test_DockerConfig_Create(t *testing.T) {
	var got *configv1beta1.DockerConfig

	want := &configv1beta1.DockerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   configv1beta1.DockerConfigSpec{},
		Status: configv1beta1.DockerConfigStatus{},
	}

	repo, err := NewDockerConfigRepository(
		log.New(),
		&mockDockerConfigGetter{
			create: func(client *mockDockerConfigClient, c *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
				got = c
				return c, nil
			},
		},
		&mockDockerConfigInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create dockerConfigs repo: %+v", err)
	}

	ctx := context.Background()
	_, err = repo.Create(ctx, want)
	if err != nil {
		t.Fatalf("cannot create dockerConfig: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected dockerConfig: %s", cmp.Diff(want, got))
	}

}

func Test_DockerConfig_Get(t *testing.T) {
	want := &configv1beta1.DockerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec:   configv1beta1.DockerConfigSpec{},
		Status: configv1beta1.DockerConfigStatus{},
	}
	informer := &mockDockerConfigInformer{}
	repo, err := NewDockerConfigRepository(
		log.New(),
		&mockDockerConfigGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create dockerConfigs repo: %+v", err)
	}

	informer.add(want)

	ctx := context.TODO()
	got, err := repo.Get(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot get dockerConfig: %+v", err)
	}

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected dockerConfig: %s", cmp.Diff(want, got))
	}
}

type DockerConfigByName []*configv1beta1.DockerConfig

func (n DockerConfigByName) Len() int {
	return len(n)
}

func (n DockerConfigByName) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n DockerConfigByName) Less(i, j int) bool {
	return strings.Compare(n[i].Name, n[j].Name) == -1
}

func Test_DockerConfig_List(t *testing.T) {
	want := make([]*configv1beta1.DockerConfig, 10)

	for i := 0; i < 10; i++ {
		want[i] = &configv1beta1.DockerConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("foo%02d", i),
				Namespace: "bar",
			},
			Spec:   configv1beta1.DockerConfigSpec{},
			Status: configv1beta1.DockerConfigStatus{},
		}
	}
	informer := &mockDockerConfigInformer{}
	repo, err := NewDockerConfigRepository(
		log.New(),
		&mockDockerConfigGetter{},
		informer,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create dockerConfigs repo: %+v", err)
	}

	for _, app := range want {
		informer.add(app)
	}

	ctx := context.TODO()
	got, err := repo.List(ctx, "bar")
	if err != nil {
		t.Fatalf("cannot get dockerConfig: %+v", err)
	}

	sort.Sort(DockerConfigByName(got))
	if !cmp.Equal(want, got) {
		t.Errorf("unexpected dockerConfig: %s", cmp.Diff(want, got))
	}
}

func Test_DockerConfig_Delete(t *testing.T) {
	var gotNamespace, gotName string

	wantName := "foo"
	wantNamespace := "bar"

	repo, err := NewDockerConfigRepository(
		log.New(),
		&mockDockerConfigGetter{
			delete: func(client *mockDockerConfigClient, name string, options *metav1.DeleteOptions) error {
				gotNamespace = client.namespace
				gotName = name
				return nil
			},
		},
		&mockDockerConfigInformer{},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("cannot create dockerConfigs repo: %+v", err)
	}
	ctx := context.TODO()
	err = repo.Delete(ctx, "bar", "foo")
	if err != nil {
		t.Fatalf("cannot delete dockerConfig: %+v", err)
	}

	if !cmp.Equal(wantName, gotName) {
		t.Errorf("unexpected name: %s", cmp.Diff(wantName, gotName))
	}
	if !cmp.Equal(wantNamespace, gotNamespace) {
		t.Errorf("unexpected namespace: %s", cmp.Diff(wantNamespace, gotNamespace))
	}
}
