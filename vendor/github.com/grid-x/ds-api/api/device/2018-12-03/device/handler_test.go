package device

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

type getF func(ctx context.Context, namespace, name string) (*corev1beta1.Device, error)
type updateF func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error)

type mockDeviceClient struct {
	get    getF
	update updateF
}

func (m *mockDeviceClient) Get(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
	return m.get(ctx, namespace, name)
}
func (m *mockDeviceClient) UpdateStatus(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
	return m.update(ctx, dev)
}

func Test_Get(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
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
				&mockDeviceClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not implemented")
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
				&mockDeviceClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Name:      name,
								Namespace: namespace,
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
							},
							Status: corev1beta1.DeviceStatus{
								LastHeartbeat: &now,
							},
						}, nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &GetResponse{
					Device: &Device{
						Metadata: api.Metadata{
							ID: "foo",
						},
						Spec: corev1beta1.DeviceSpec{
							Serialnumber: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
						},
						Status: corev1beta1.DeviceStatus{
							LastHeartbeat: &now,
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

func Test_Update(t *testing.T) {
	n := time.Now()
	now := n.Format(time.RFC3339)
	now2m := n.Add(-2 * time.Minute).Format(time.RFC3339)
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
				&mockDeviceClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not implemented")
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
			},
			input: UpdateRequest{
				Status: &corev1beta1.DeviceStatus{
					LastHeartbeat: &now,
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
				&mockDeviceClient{
					get: func(ctx context.Context, namespace, name string) (*corev1beta1.Device, error) {
						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Name:      name,
								Namespace: namespace,
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber: "ff8a5861-824b-4246-9457-94f04d665d7b",
							},
							Status: corev1beta1.DeviceStatus{
								LastHeartbeat: &now2m,
							},
						}, nil
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return dev, nil
					},
				},
			},
			input: UpdateRequest{
				Status: &corev1beta1.DeviceStatus{
					LastHeartbeat: &now,
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Device: &Device{
						Metadata: api.Metadata{
							ID: "foo",
						},
						Spec: corev1beta1.DeviceSpec{
							Serialnumber: "ff8a5861-824b-4246-9457-94f04d665d7b",
						},
						Status: corev1beta1.DeviceStatus{
							LastHeartbeat: &now,
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
