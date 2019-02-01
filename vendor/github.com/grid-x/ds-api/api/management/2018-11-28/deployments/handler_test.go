package deployments

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
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

type deployCreateF func(context.Context, *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
type deployUpdateF func(context.Context, *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
type deployDeleteF func(context.Context, string, string) error

type deployGetF func(context.Context, string, string) (*appsv1beta1.DeviceDeployment, error)
type deployListF func(context.Context, string) ([]*appsv1beta1.DeviceDeployment, error)

type mockK8sDeploy struct {
	create deployCreateF
	get    deployGetF
	list   deployListF
	update deployUpdateF
	delete deployDeleteF
}

func (m *mockK8sDeploy) Create(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return m.create(ctx, deploy)
}

func (m *mockK8sDeploy) Get(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
	return m.get(ctx, accountID, uuid)
}

func (m *mockK8sDeploy) List(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
	return m.list(ctx, accountID)
}

func (m *mockK8sDeploy) Update(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
	return m.update(ctx, deploy)
}

func (m *mockK8sDeploy) Delete(ctx context.Context, namespace, name string) error {
	return m.delete(ctx, namespace, name)
}

type appGetF func(context.Context, string, string) (*appsv1beta1.DeviceApplication, error)

type mockAppRepo struct {
	get appGetF
}

func (m *mockAppRepo) Get(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
	return m.get(ctx, accountID, name)
}

func newReq(id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"deploymentID": id,
	})

	return req
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
		{ // Test missing app name
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "testcontainer",
									Image: "test:1234567",
								},
							},
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
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test missing container spec
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{},
					},
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
				&mockK8sDeploy{
					create: func(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
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
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Specified app not found
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDeploy{
					create: func(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("no application found")
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("no application found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDeploy{
					create: func(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
						return deploy, nil
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
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
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "testcontainer",
									Image: "test:1234567",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					Deployment: &Deployment{
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						Spec: appsv1beta1.DeviceDeploymentSpec{
							App: "monitoring",
							Selector: appsv1beta1.Selector{
								MatchByLabels: map[string]string{
									"foo": "bar",
								},
							},
							Template: appsv1beta1.PodTemplate{
								Spec: corev1beta1.PodConfig{
									Containers: []corev1beta1.Container{
										{
											Name:  "testcontainer",
											Image: "test:1234567",
										},
									},
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
				&mockK8sDeploy{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("not yet implemented")
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
				&mockK8sDeploy{
					list: func(ctx context.Context, accountID string) ([]*appsv1beta1.DeviceDeployment, error) {
						return []*appsv1beta1.DeviceDeployment{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "ebdd58ad-3094-45cd-889e-910d6ce5e9e8",
									Namespace: accountID,
								},
								Spec:   appsv1beta1.DeviceDeploymentSpec{},
								Status: appsv1beta1.DeviceDeploymentStatus{},
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
					Deployments: []*Deployment{
						{
							Metadata: api.Metadata{
								ID: "@ignore",
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
			got, err := svc.List(tc.req)

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

func Test_Get(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDeploy{
					get: func(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("not yet implemented")
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
			req: newReq("587e7ce4-eed3-4dca-8e29-349b91421b45", t),
			injections: []interface{}{
				&mockK8sDeploy{
					get: func(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
						return &appsv1beta1.DeviceDeployment{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      uuid,
							},
							Spec:   appsv1beta1.DeviceDeploymentSpec{},
							Status: appsv1beta1.DeviceDeploymentStatus{},
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
					Deployment: &Deployment{
						Metadata: api.Metadata{
							ID: "@ignore",
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

			if !cmp.Equal(tc.want, got, testutils.CmpWithIgnore("ID")) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got, testutils.CmpWithIgnore("ID")))
			}
		})
	}
}

func Test_Update_Validate(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateRequest

		wantErr bool
		want    error
	}{
		{ // Test empty request
			req:        &http.Request{},
			injections: []interface{}{},
			input:      UpdateRequest{},
			wantErr:    true,
			want:       nil,
		},
		{ // Test missing app name
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "testcontainer",
									Image: "test:1234567",
								},
							},
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
			input: UpdateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test missing container spec
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{},
					},
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

func Test_Update(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateRequest

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDeploy{
					get: func(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("not found")
					},
					update: func(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{},

			wantErr: true,
			want:    nil,
		}, {
			req: newReq("ca009f17-12f9-43c7-98bc-b6863bbc787a", t),
			injections: []interface{}{
				&mockK8sDeploy{
					get: func(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
						if uuid != "ca009f17-12f9-43c7-98bc-b6863bbc787a" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}
						return &appsv1beta1.DeviceDeployment{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      uuid,
							},
							Spec: appsv1beta1.DeviceDeploymentSpec{
								App: "monitoring",
								Selector: appsv1beta1.Selector{
									MatchByLabels: map[string]string{
										"foo": "bar",
									},
								},
								Template: appsv1beta1.PodTemplate{
									Spec: corev1beta1.PodConfig{
										Containers: []corev1beta1.Container{
											{
												Name:  "testcontainer",
												Image: "test:1234567",
											},
										},
									},
								},
							},
							Status: appsv1beta1.DeviceDeploymentStatus{},
						}, nil
					},
					update: func(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error) {
						if deploy.Name != "ca009f17-12f9-43c7-98bc-b6863bbc787a" {
							return nil, fmt.Errorf("unexpected deployment ID")
						}
						return deploy, nil
					},
				},
				&mockAppRepo{
					get: func(ctx context.Context, accountID, name string) (*appsv1beta1.DeviceApplication, error) {
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
			input: UpdateRequest{
				Spec: &appsv1beta1.DeviceDeploymentSpec{
					App: "monitoring-new",
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "baz",
						},
					},
					Template: appsv1beta1.PodTemplate{
						Spec: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Name:  "testcontainer",
									Image: "test:99",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Deployment: &Deployment{
						Metadata: api.Metadata{
							ID: "ca009f17-12f9-43c7-98bc-b6863bbc787a",
						},
						Spec: appsv1beta1.DeviceDeploymentSpec{
							App: "monitoring-new",
							Selector: appsv1beta1.Selector{
								MatchByLabels: map[string]string{
									"foo": "baz",
								},
							},
							Template: appsv1beta1.PodTemplate{
								Spec: corev1beta1.PodConfig{
									Containers: []corev1beta1.Container{
										{
											Name:  "testcontainer",
											Image: "test:99",
										},
									},
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
			got, err := svc.Update(tc.req, tc.input)

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
				&mockK8sDeploy{
					delete: func(ctx context.Context, namespace, name string) error {
						return fmt.Errorf("not yet implemented")
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
			req: newReq("587e7ce4-eed3-4dca-8e29-349b91421b45", t),
			injections: []interface{}{
				&mockK8sDeploy{
					delete: func(ctx context.Context, namespace, name string) error {
						return nil
					},
					get: func(ctx context.Context, accountID, uuid string) (*appsv1beta1.DeviceDeployment, error) {
						return &appsv1beta1.DeviceDeployment{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      uuid,
							},
							Spec:   appsv1beta1.DeviceDeploymentSpec{},
							Status: appsv1beta1.DeviceDeploymentStatus{},
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
			want:    &encoding.Response{Payload: &DeleteResponse{}},
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
