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
	deviceID, accountID string
}

func (m *mockAuthProvider) AccountIDFromContext(context.Context) (string, error) {
	return m.accountID, nil
}

func (m *mockAuthProvider) DeviceIDFromContext(context.Context) (string, error) {
	return m.deviceID, nil
}

type getFunc func(context.Context, string, string) (*corev1beta1.DevicePod, error)
type listFunc func(context.Context, string, string) ([]*corev1beta1.DevicePod, error)
type updateFunc func(context.Context, *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	get    getFunc
	list   listFunc
	update updateFunc
}

func (m *mockPodClient) ListByDeviceID(ctx context.Context, namespace string, deviceID string) ([]*corev1beta1.DevicePod, error) {
	return m.list(ctx, namespace, deviceID)
}
func (m *mockPodClient) UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return m.update(ctx, pod)
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
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					list: func(ctx context.Context, namespace, deviceID string) ([]*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					list: func(ctx context.Context, namespace, deviceID string) ([]*corev1beta1.DevicePod, error) {
						return []*corev1beta1.DevicePod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: namespace,
									Name:      "894e3aa3-8beb-4e79-9785-3bd35fa671bc",
								},
								Spec:   corev1beta1.DevicePodSpec{},
								Status: corev1beta1.DevicePodStatus{},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: namespace,
									Name:      "e76f475d-4d30-4d5e-a10f-2a99afb24e87",
								},
								Spec:   corev1beta1.DevicePodSpec{},
								Status: corev1beta1.DevicePodStatus{},
							},
						}, nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &ListResponse{
					Pods: []*Pod{
						{
							Metadata: api.Metadata{
								ID: "894e3aa3-8beb-4e79-9785-3bd35fa671bc",
							},
						},
						{
							Metadata: api.Metadata{
								ID: "e76f475d-4d30-4d5e-a10f-2a99afb24e87",
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

func Test_Update(t *testing.T) {
	now := metav1.Now()
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
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					update: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
				},
			},
			input: UpdateRequest{
				Status: &corev1beta1.DevicePodStatus{
					StartTime: &now,
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: newReq("c91ffe91-e44b-4fa3-946a-b153d9a9ecbb", t),
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					update: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
						return pod, nil
					},
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error) {
						return &corev1beta1.DevicePod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: namespace,
								Name:      "c91ffe91-e44b-4fa3-946a-b153d9a9ecbb",
							},
							Spec:   corev1beta1.DevicePodSpec{},
							Status: corev1beta1.DevicePodStatus{},
						}, nil
					},
				},
			},
			input: UpdateRequest{
				Status: &corev1beta1.DevicePodStatus{
					StartTime: &now,
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Pod: &Pod{
						Metadata: api.Metadata{
							ID: "c91ffe91-e44b-4fa3-946a-b153d9a9ecbb",
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
			got, err := svc.Update(tc.req, tc.input)

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
