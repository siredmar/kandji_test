package application

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
)

type mockAuthProvider struct {
	f func(context.Context) (string, error)
}

func (m *mockAuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	return m.f(ctx)
}

type createAppF func(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error)
type getAppF func(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceApplication, error)
type listAppsF func(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceApplication, error)
type deleteAppF func(ctx context.Context, namespace, name string) error

type mockAppRepo struct {
	create createAppF
	get    getAppF
	list   listAppsF
	delete deleteAppF
}

func (m *mockAppRepo) Create(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
	return m.create(ctx, application)
}

func (m *mockAppRepo) Get(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
	return m.get(ctx, accountID, id)
}

func (m *mockAppRepo) List(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceApplication, error) {
	return m.list(ctx, accountID)
}

func (m *mockAppRepo) Delete(ctx context.Context, accountID, id string) error {
	return m.delete(ctx, accountID, id)
}

type deployListF func(context.Context, string) ([]*appsv1beta1.DeviceDeployment, error)

type mockDeploymentRepo struct {
	list deployListF
}

func (m *mockDeploymentRepo) List(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
	return m.list(ctx, accountID)
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
		{ // Test empty name
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Name: "",
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
				&mockAppRepo{
					create: func(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("not implemented yet")
					},
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("not implemented yet")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input:   CreateRequest{},
			wantErr: true,
			want:    nil,
		},
		{ // Test already existing application
			req: &http.Request{},
			injections: []interface{}{
				&mockAppRepo{
					create: func(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
						return application, nil
					},
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return &appsv1beta1.DeviceApplication{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "monitoring",
								Namespace: accountID,
							},
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Name: "monitoring",
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAppRepo{
					create: func(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error) {
						return application, nil
					},
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Name: "monitoring",
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					Application: Application{
						Metadata: api.Metadata{
							ID: "monitoring",
						},
						Name: "monitoring",
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

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func newReq(id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "test.com", nil)
	if err != nil {
		t.Fatalf("cannot create request: %+v", err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"applicationID": id,
	})
	return req
}

func Test_Get(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: newReq("7dc71d66-41cf-4b72-b058-3aac2c31532e", t),
			injections: []interface{}{
				&mockAppRepo{
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("not found")
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
			req: newReq("7dc71d66-41cf-4b72-b058-3aac2c31532e", t),
			injections: []interface{}{
				&mockAppRepo{
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return &appsv1beta1.DeviceApplication{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "monitoring",
								Namespace: accountID,
							},
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
					Application: Application{
						Metadata: api.Metadata{
							ID: "monitoring",
						},
						Name: "monitoring",
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
				&mockAppRepo{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("internal error")
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
			req: &http.Request{},
			injections: []interface{}{
				&mockAppRepo{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceApplication, error) {
						return []*appsv1beta1.DeviceApplication{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "monitoring",
									Namespace: accountID,
								},
								Spec:   appsv1beta1.DeviceApplicationSpec{},
								Status: appsv1beta1.DeviceApplicationStatus{},
							},
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
				Payload: &ListResponse{
					Applications: []Application{
						{
							Metadata: api.Metadata{
								ID: "monitoring",
							},
							Name: "monitoring",
						},
					},
				},
			},
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAppRepo{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceApplication, error) {
						return []*appsv1beta1.DeviceApplication{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "monitoring",
									Namespace: accountID,
								},
								Spec:   appsv1beta1.DeviceApplicationSpec{},
								Status: appsv1beta1.DeviceApplicationStatus{},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "envscan",
									Namespace: accountID,
								},
								Spec:   appsv1beta1.DeviceApplicationSpec{},
								Status: appsv1beta1.DeviceApplicationStatus{},
							},
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
				Payload: &ListResponse{
					Applications: []Application{
						{
							Metadata: api.Metadata{
								ID: "envscan",
							},
							Name: "envscan",
						},
						{
							Metadata: api.Metadata{
								ID: "monitoring",
							},
							Name: "monitoring",
						},
					},
				},
			},
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
				&mockAppRepo{
					delete: func(ctx context.Context, accountID, id string) error {
						return fmt.Errorf("internal error")
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
			req: newReq("monitoring", t),
			injections: []interface{}{
				&mockDeploymentRepo{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
						return []*appsv1beta1.DeviceDeployment{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "ebdd58ad-3094-45cd-889e-910d6ce5e9e8",
									Namespace: accountID,
								},
								Spec: appsv1beta1.DeviceDeploymentSpec{
									App: "monitoring",
								},
								Status: appsv1beta1.DeviceDeploymentStatus{},
							},
						}, nil
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return &appsv1beta1.DeviceApplication{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "monitoring",
								Namespace: accountID,
							},
						}, nil
					},
					delete: func(ctx context.Context, accountID, id string) error {
						return fmt.Errorf("bound to deployments")
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
			req: newReq("monitoring", t),
			injections: []interface{}{
				&mockDeploymentRepo{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
						return []*appsv1beta1.DeviceDeployment{}, nil
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, id string) (*appsv1beta1.DeviceApplication, error) {
						return &appsv1beta1.DeviceApplication{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "monitoring",
								Namespace: accountID,
							},
						}, nil
					},
					delete: func(ctx context.Context, accountID, id string) error {
						return nil
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
