package maintenance

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	testutils "github.com/grid-x/ds-api/pkg/testing"
)

type mockAuthProvider struct {
	f func(context.Context) (string, error)
}

func (m *mockAuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	return m.f(ctx)
}

type createF func(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error)
type updateF func(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error)
type deleteF func(ctx context.Context, namespace, name string) error
type listF func(ctx context.Context, namespace string) ([]*mainv1beta1.MaintenanceTask, error)
type getF func(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error)

type mockMaintenanceTaskClient struct {
	create createF
	update updateF
	delete deleteF
	list   listF
	get    getF
}

func (m *mockMaintenanceTaskClient) Create(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	return m.create(ctx, task)
}

func (m *mockMaintenanceTaskClient) Update(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
	return m.update(ctx, task)
}

func (m *mockMaintenanceTaskClient) Delete(ctx context.Context, namespace, name string) error {
	return m.delete(ctx, namespace, name)
}

func (m *mockMaintenanceTaskClient) List(ctx context.Context, namespace string) ([]*mainv1beta1.MaintenanceTask, error) {
	return m.list(ctx, namespace)
}

func (m *mockMaintenanceTaskClient) Get(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error) {
	return m.get(ctx, namespace, name)
}

func newReq(method string, id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	return mux.SetURLVars(req, map[string]string{
		"taskID": id,
	})
}

func Test_Create_Validate(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      CreateRequest

		wantErr bool
		want    error
	}{
		{ // Test empty request
			req:        &http.Request{},
			injections: []interface{}{},
			input:      CreateRequest{},
			wantErr:    true,
			want:       nil,
		},
		{ // Test missing type
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &mainv1beta1.MaintenanceTaskSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test missing MatchByLabels selector
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &mainv1beta1.MaintenanceTaskSpec{
					Type:     mainv1beta1.MaintenanceTaskTypeRestart,
					Selector: appsv1beta1.Selector{},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tc.input.Validate()

			if !tc.wantErr && got != nil {
				t.Fatalf("unexpected error: %+v", got)
			} else if tc.wantErr && got == nil {
				t.Fatal("expected error but did not get one")
			}
		})
	}
}

func Test_Create(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      CreateRequest

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					create: func(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: &mainv1beta1.MaintenanceTaskSpec{
					Type: mainv1beta1.MaintenanceTaskTypeRestart,
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					create: func(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error) {
						return task, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: &mainv1beta1.MaintenanceTaskSpec{
					Type: mainv1beta1.MaintenanceTaskTypeRestart,
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					Task: &Task{
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						Spec: mainv1beta1.MaintenanceTaskSpec{
							Type: mainv1beta1.MaintenanceTaskTypeRestart,
							Selector: appsv1beta1.Selector{
								MatchByLabels: map[string]string{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Create(tc.req, tc.input)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got, testutils.CmpWithIgnore("ID")) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got, testutils.CmpWithIgnore("ID")))
			}
		})
	}
}

func Test_List(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					list: func(ctx context.Context, namespace string) ([]*mainv1beta1.MaintenanceTask, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.List(tc.req)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func Test_Get(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: newReq(http.MethodGet, "293540a0-d43f-4749-8835-1f30d3b12279", t),
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					get: func(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: newReq(http.MethodGet, "293540a0-d43f-4749-8835-1f30d3b12279", t),
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					get: func(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error) {
						return &mainv1beta1.MaintenanceTask{
							ObjectMeta: metav1.ObjectMeta{
								Name:      name,
								Namespace: namespace,
							},
							Spec: mainv1beta1.MaintenanceTaskSpec{
								Type: mainv1beta1.MaintenanceTaskTypeRestart,
								Selector: appsv1beta1.Selector{
									MatchByLabels: map[string]string{
										"foo": "bar",
									},
								},
							},
							Status: mainv1beta1.MaintenanceTaskStatus{},
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &GetResponse{
					Task: &Task{
						Metadata: api.Metadata{
							ID: "293540a0-d43f-4749-8835-1f30d3b12279",
						},
						Spec: mainv1beta1.MaintenanceTaskSpec{
							Type: mainv1beta1.MaintenanceTaskTypeRestart,
							Selector: appsv1beta1.Selector{
								MatchByLabels: map[string]string{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Get(tc.req)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					delete: func(ctx context.Context, namespace, name string) error {
						return fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: newReq(http.MethodDelete, "d93abd18-97e3-4495-ba45-f809ee368347", t),
			injections: []interface{}{
				&mockMaintenanceTaskClient{
					delete: func(ctx context.Context, namespace, name string) error {
						return nil
					},
					get: func(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error) {
						return &mainv1beta1.MaintenanceTask{
							ObjectMeta: metav1.ObjectMeta{
								Name:      name,
								Namespace: namespace,
							},
							Spec: mainv1beta1.MaintenanceTaskSpec{
								Type: mainv1beta1.MaintenanceTaskTypeRestart,
								Selector: appsv1beta1.Selector{
									MatchByLabels: map[string]string{
										"foo": "bar",
									},
								},
							},
							Status: mainv1beta1.MaintenanceTaskStatus{},
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &DeleteResponse{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Delete(tc.req)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
