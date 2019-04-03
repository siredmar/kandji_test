package dockerconfigs

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
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

type dockerConfigCreateF func(context.Context, *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error)
type dockerConfigDeleteF func(context.Context, string, string) error

type dockerConfigGetF func(context.Context, string, string) (*configv1beta1.DockerConfig, error)
type dockerConfigListF func(context.Context, string) ([]*configv1beta1.DockerConfig, error)

type mockK8sDockerConfig struct {
	create dockerConfigCreateF
	get    dockerConfigGetF
	list   dockerConfigListF
	delete dockerConfigDeleteF
}

func (m *mockK8sDockerConfig) Create(ctx context.Context, dockerConfig *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
	return m.create(ctx, dockerConfig)
}

func (m *mockK8sDockerConfig) Get(ctx context.Context, accountID, uuid string) (*configv1beta1.DockerConfig, error) {
	return m.get(ctx, accountID, uuid)
}

func (m *mockK8sDockerConfig) List(ctx context.Context, accountID string) ([]*configv1beta1.DockerConfig, error) {
	return m.list(ctx, accountID)
}

func (m *mockK8sDockerConfig) Delete(ctx context.Context, namespace, name string) error {
	return m.delete(ctx, namespace, name)
}

func newReq(id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"dockerConfigID": id,
	})

	return req
}

func Test_CreateRequest_Validate(t *testing.T) {
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
		{ // Test missing registry
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &configv1beta1.DockerConfigSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Registry: "",
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test missing creds
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &configv1beta1.DockerConfigSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Registry:    "https://630781358184.dkr.ecr.eu-central-1.amazonaws.com",
					Credentials: configv1beta1.DockerConfigCredentails{},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test valid obne
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: &configv1beta1.DockerConfigSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Registry: "https://630781358184.dkr.ecr.eu-central-1.amazonaws.com",
					Credentials: configv1beta1.DockerConfigCredentails{
						AWS: &configv1beta1.AWSCredentialProvider{
							AccessKey:       "foo",
							SecretAccessKey: "foo",
							Region:          "foo",
						},
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
				&mockK8sDockerConfig{
					create: func(ctx context.Context, deploy *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
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
				Spec: &configv1beta1.DockerConfigSpec{
					Registry: "foo",
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDockerConfig{
					create: func(ctx context.Context, deploy *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error) {
						return deploy, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: &configv1beta1.DockerConfigSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Registry: "https://630781358184.dkr.ecr.eu-central-1.amazonaws.com",
					Credentials: configv1beta1.DockerConfigCredentails{
						AWS: &configv1beta1.AWSCredentialProvider{
							AccessKey:       "foo",
							SecretAccessKey: "foo",
							Region:          "foo",
						},
					},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					DockerConfig: &DockerConfig{
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						Spec: configv1beta1.DockerConfigSpec{
							Selector: appsv1beta1.Selector{
								MatchByLabels: map[string]string{
									"foo": "bar",
								},
							},
							Registry: "https://630781358184.dkr.ecr.eu-central-1.amazonaws.com",
							Credentials: configv1beta1.DockerConfigCredentails{
								AWS: &configv1beta1.AWSCredentialProvider{
									AccessKey:       "foo",
									SecretAccessKey: "foo",
									Region:          "foo",
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
				&mockK8sDockerConfig{
					get: func(ctx context.Context, accountID, uuid string) (*configv1beta1.DockerConfig, error) {
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
			req: newReq("587e7ce4-eed3-4dca-8e29-349b91421b45", t),
			injections: []interface{}{
				&mockK8sDockerConfig{
					get: func(ctx context.Context, accountID, uuid string) (*configv1beta1.DockerConfig, error) {
						return &configv1beta1.DockerConfig{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      uuid,
							},
							Spec:   configv1beta1.DockerConfigSpec{},
							Status: configv1beta1.DockerConfigStatus{},
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
					DockerConfig: &DockerConfig{
						Metadata: api.Metadata{
							ID: "587e7ce4-eed3-4dca-8e29-349b91421b45",
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
				&mockK8sDockerConfig{
					list: func(ctx context.Context, accountID string) ([]*configv1beta1.DockerConfig, error) {
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
			req: &http.Request{},
			injections: []interface{}{
				&mockK8sDockerConfig{
					list: func(ctx context.Context, accountID string) ([]*configv1beta1.DockerConfig, error) {
						return []*configv1beta1.DockerConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "587e7ce4-eed3-4dca-8e29-349b91421b45",
									Namespace: accountID,
								},
								Spec:   configv1beta1.DockerConfigSpec{},
								Status: configv1beta1.DockerConfigStatus{},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "ebdd58ad-3094-45cd-889e-910d6ce5e9e8",
									Namespace: accountID,
								},
								Spec:   configv1beta1.DockerConfigSpec{},
								Status: configv1beta1.DockerConfigStatus{},
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
					DockerConfigs: []*DockerConfig{
						{
							Metadata: api.Metadata{
								ID: "587e7ce4-eed3-4dca-8e29-349b91421b45",
							},
						},
						{
							Metadata: api.Metadata{
								ID: "ebdd58ad-3094-45cd-889e-910d6ce5e9e8",
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
				&mockK8sDockerConfig{
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
			req: newReq("587e7ce4-eed3-4dca-8e29-349b91421b45", t),
			injections: []interface{}{
				&mockK8sDockerConfig{
					delete: func(ctx context.Context, namespace, name string) error {
						return nil
					},
					get: func(ctx context.Context, accountID, uuid string) (*configv1beta1.DockerConfig, error) {
						return &configv1beta1.DockerConfig{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      uuid,
							},
							Spec:   configv1beta1.DockerConfigSpec{},
							Status: configv1beta1.DockerConfigStatus{},
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
