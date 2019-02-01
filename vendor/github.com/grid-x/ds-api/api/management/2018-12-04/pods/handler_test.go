package pods

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
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

type listF func(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error)
type getF func(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	list listF
	get  getF
}

func (m *mockPodClient) List(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error) {
	return m.list(ctx, namespace)
}
func (m *mockPodClient) Get(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
	return m.get(ctx, namespace, name)
}

func newReq(id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	return mux.SetURLVars(req, map[string]string{
		"podID": id,
	})
}

func Test_List(t *testing.T) {
	now := metav1.Now()
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockPodClient{
					list: func(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error) {
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
				&mockPodClient{
					list: func(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error) {
						return []*corev1beta1.DevicePod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: namespace,
									Name:      "04449e31-5818-40bc-8537-cacf7a4e9864",
								},
								Spec: corev1beta1.DevicePodSpec{
									DeviceID: "13d49cd8-3c17-4f73-9ccc-c898a60f842b",
								},
								Status: corev1beta1.DevicePodStatus{
									StartTime: &now,
								},
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
					Pods: []*Pod{
						{
							Metadata: api.Metadata{
								ID: "04449e31-5818-40bc-8537-cacf7a4e9864",
							},
							Spec: corev1beta1.DevicePodSpec{
								DeviceID: "13d49cd8-3c17-4f73-9ccc-c898a60f842b",
							},
							Status: corev1beta1.DevicePodStatus{
								StartTime: &now,
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

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func Test_Get(t *testing.T) {
	now := metav1.Now()
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
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
			req: newReq("7b509d9a-86f2-4f78-88ee-3e130bccb456", t),
			injections: []interface{}{
				&mockPodClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
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
			req: newReq("7b509d9a-86f2-4f78-88ee-3e130bccb456", t),
			injections: []interface{}{
				&mockPodClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
						return &corev1beta1.DevicePod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      name,
								Namespace: namespace,
							},
							Spec: corev1beta1.DevicePodSpec{
								DeviceID: "9fed5c90-0f75-4c3d-81b3-07f9df291ac2",
							},
							Status: corev1beta1.DevicePodStatus{
								StartTime: &now,
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
					Pod: &Pod{
						Metadata: api.Metadata{
							ID: "7b509d9a-86f2-4f78-88ee-3e130bccb456",
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "9fed5c90-0f75-4c3d-81b3-07f9df291ac2",
						},
						Status: corev1beta1.DevicePodStatus{
							StartTime: &now,
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
