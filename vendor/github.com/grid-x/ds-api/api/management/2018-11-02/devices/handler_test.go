package devices

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/ssh"
	testutils "github.com/grid-x/ds-api/pkg/testing"
)

func newReqWithDeviceID(id string) *http.Request {
	req, _ := http.NewRequest("GET", "/devices/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"deviceID": id})
	return req
}

type mockUUIDGen func() uuid.UUID

type mockAuthProvider struct {
	f func(context.Context) (string, error)
}

func (m *mockAuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	return m.f(ctx)
}

type deviceCreateF func(context.Context, *corev1beta1.Device) (*corev1beta1.Device, error)
type deviceGetF func(context.Context, string, string) (*corev1beta1.Device, error)
type deviceGetBySerialnumberF func(context.Context, string) (*corev1beta1.Device, error)
type deviceListF func(context.Context, string) ([]*corev1beta1.Device, error)
type deviceUpdateF func(context.Context, *corev1beta1.Device) (*corev1beta1.Device, error)
type deviceDeleteF func(context.Context, string, string) error

type mockDeviceClient struct {
	create            deviceCreateF
	get               deviceGetF
	getBySerialnumber deviceGetBySerialnumberF
	list              deviceListF
	update            deviceUpdateF
	delete            deviceDeleteF
}

func (c *mockDeviceClient) Create(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
	return c.create(ctx, dev)
}

func (c *mockDeviceClient) Get(ctx context.Context, accountID, id string) (*corev1beta1.Device, error) {
	return c.get(ctx, accountID, id)
}

func (c *mockDeviceClient) GetBySerialnumber(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
	return c.getBySerialnumber(ctx, serialnumber)
}

func (c *mockDeviceClient) List(ctx context.Context, accountID string) ([]*corev1beta1.Device, error) {
	return c.list(ctx, accountID)
}

func (c *mockDeviceClient) Update(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
	return c.update(ctx, dev)
}

func (c *mockDeviceClient) Delete(ctx context.Context, accountID, id string) error {
	return c.delete(ctx, accountID, id)
}

type podCreateF func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	create podCreateF
}

func (m *mockPodClient) Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return m.create(ctx, pod)
}

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
}

func mkStringMap(m map[string]string) *map[string]string {
	r := new(map[string]string)
	*r = m
	return r
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
		{ // Test empty serialnumber
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: DeviceSpec{
					Serialnumber: "",
				},
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test incorrect maintenance window
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Spec: DeviceSpec{
					Serialnumber:      "foo-bar-123",
					MaintenanceWindow: mkString(":23:00-Tue:02:00"),
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
				&mockDeviceClient{
					getBySerialnumber: func(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
						return nil, nil
					},
					create: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return dev, nil
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
		{ // Test already existing serial number
			req: &http.Request{},
			injections: []interface{}{
				&mockDeviceClient{
					getBySerialnumber: func(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "account-ef14066c-e7d7-4d38-91a6-06ba7231b633",
								Name:      "f61f78fb-c52e-4df9-8b1a-2219b2188bd4",
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "foo-bar-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
						}, nil
					},
					create: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: DeviceSpec{
					Serialnumber: "foo-bar-123",
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockDeviceClient{
					getBySerialnumber: func(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not found")
					},
					create: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: DeviceSpec{
					Serialnumber:      "foo-bar-123",
					MaintenanceWindow: mkString("Mon:23:00-Tue:02:00"),
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					Device: &Device{
						Status: DeviceStatus{},
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						Spec: DeviceSpec{
							Serialnumber:      "foo-bar-123",
							MaintenanceWindow: mkString("Mon:23:00-Tue:02:00"),
						},
					},
				},
			},
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockDeviceClient{
					getBySerialnumber: func(ctx context.Context, serialnumber string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not found")
					},
					create: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				Spec: DeviceSpec{
					Serialnumber:      "foo-bar-123",
					MACAddress:        mkString("00:00:00:00:00:00:00:00"),
					MaintenanceWindow: mkString("Mon:23:00-Tue:02:00"),
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					Device: &Device{
						Status: DeviceStatus{},
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						Spec: DeviceSpec{
							Serialnumber:      "foo-bar-123",
							MACAddress:        mkString("00:00:00:00:00:00:00:00"),
							MaintenanceWindow: mkString("Mon:23:00-Tue:02:00"),
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
				&mockDeviceClient{
					list: func(ctx context.Context, accountID string) ([]*corev1beta1.Device, error) {
						return nil, nil
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
					Devices: []*Device{},
				},
			},
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockDeviceClient{
					list: func(ctx context.Context, accountID string) ([]*corev1beta1.Device, error) {
						return []*corev1beta1.Device{
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: accountID,
									Name:      "f61f78fb-c52e-4df9-8b1a-2219b2188bd4",
								},
								Spec: corev1beta1.DeviceSpec{
									Serialnumber:      "device-5678",
									MaintenanceWindow: "Sun:04:00-Sun:06:00",
								},
								Status: corev1beta1.DeviceStatus{},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: accountID,
									Name:      "ef14066c-e7d7-4d38-91a6-06ba7231b633",
								},
								Spec: corev1beta1.DeviceSpec{
									Serialnumber:      "device-1234",
									MaintenanceWindow: "Sun:04:00-Sun:06:00",
								},
								Status: corev1beta1.DeviceStatus{},
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
					Devices: []*Device{
						{
							Metadata: api.Metadata{
								ID: "ef14066c-e7d7-4d38-91a6-06ba7231b633",
							},
							Spec: DeviceSpec{
								Serialnumber:      "device-1234",
								MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
							},
						},
						{
							Metadata: api.Metadata{
								ID: "f61f78fb-c52e-4df9-8b1a-2219b2188bd4",
							},
							Spec: DeviceSpec{
								Serialnumber:      "device-5678",
								MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
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
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, deviceID string) (*corev1beta1.Device, error) {
						if deviceID != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}

						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      deviceID,
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "serial-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
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
					Device: &Device{
						Metadata: api.Metadata{
							ID: "2a598bf6-7591-45f1-979f-1fd07f97704c",
						},
						Spec: DeviceSpec{
							Serialnumber:      "serial-123",
							MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
						},
						Status: DeviceStatus{},
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
		{ // Test incorrect maintenance window
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				Spec: UpdateSpec{
					MaintenanceWindow: mkString(":23:00-Tue:02:00"),
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
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, id string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not found")
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input:   UpdateRequest{},
			wantErr: true,
			want:    nil,
		},
		{
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, id string) (*corev1beta1.Device, error) {
						if id != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}

						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      "2a598bf6-7591-45f1-979f-1fd07f97704c",
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "serial-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
						}, nil
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						if dev.Name != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{
				Spec: UpdateSpec{
					MACAddress: mkString("foobar"),
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Device: &Device{
						Metadata: api.Metadata{
							ID: "2a598bf6-7591-45f1-979f-1fd07f97704c",
						},
						Spec: DeviceSpec{
							Serialnumber:      "serial-123",
							MACAddress:        mkString("foobar"),
							MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
						},
					},
				},
			},
		},
		{
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, id string) (*corev1beta1.Device, error) {
						if id != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}

						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      "2a598bf6-7591-45f1-979f-1fd07f97704c",
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "serial-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
						}, nil
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						if dev.Name != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{
				Metadata: api.UpdateMetadata{
					Labels: map[string]string{"gridx.de/channel": "stable"},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Device: &Device{
						Metadata: api.Metadata{
							ID:     "2a598bf6-7591-45f1-979f-1fd07f97704c",
							Labels: map[string]string{"gridx.de/channel": "stable"},
						},
						Spec: DeviceSpec{
							Serialnumber:      "serial-123",
							MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
						},
					},
				},
			},
		},
		{
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, id string) (*corev1beta1.Device, error) {
						if id != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}

						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      "2a598bf6-7591-45f1-979f-1fd07f97704c",
								Labels:    map[string]string{"gridx.de/channel": "stable"},
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "serial-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
						}, nil
					},
					update: func(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error) {
						if dev.Name != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}
						return dev, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{
				Metadata: api.UpdateMetadata{
					Labels: map[string]string{"gridx.de/channel-": ""},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Device: &Device{
						Metadata: api.Metadata{
							ID:     "2a598bf6-7591-45f1-979f-1fd07f97704c",
							Labels: map[string]string{},
						},
						Spec: DeviceSpec{
							Serialnumber:      "serial-123",
							MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
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

func Test_Delete(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, deviceID string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not found")
					},
					delete: func(ctx context.Context, accountID, id string) error {
						return fmt.Errorf("not found")
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
			req: newReqWithDeviceID("2a598bf6-7591-45f1-979f-1fd07f97704c"),
			injections: []interface{}{
				&mockDeviceClient{
					get: func(ctx context.Context, accountID, deviceID string) (*corev1beta1.Device, error) {
						if deviceID != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return nil, fmt.Errorf("got unexpected deviceID")
						}

						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: accountID,
								Name:      deviceID,
							},
							Spec: corev1beta1.DeviceSpec{
								Serialnumber:      "serial-123",
								MaintenanceWindow: "Sun:04:00-Sun:06:00",
							},
							Status: corev1beta1.DeviceStatus{},
						}, nil
					},
					delete: func(ctx context.Context, accountID, id string) error {
						if id != "2a598bf6-7591-45f1-979f-1fd07f97704c" {
							return fmt.Errorf("got unexpected id: %s", id)
						}

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

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestSSHCreate(t *testing.T) {
	injections := []interface{}{
		&mockAuthProvider{
			f: func(ctx context.Context) (string, error) {
				return "default", nil
			},
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				if pod.Name != "894e3aa3-8beb-4e79-9785-3bd35fa671bc" {
					t.Fatalf("Wrong pod name: %s", pod.Name)
				}
				if pod.Spec.Config.Network != "Host" {
					t.Fatalf("No host network in podconfig %s", pod.Spec.Config.Network)
				}
				if pod.Spec.Config.Containers[0].Image != "ds-ssh-agent:latest" {
					t.Fatalf("Wrong image name: %s", pod.Spec.Config.Containers[0].Image)
				}
				return pod, nil
			},
		},
	}

	// Some UUIDs to use within the test...
	deviceUUID := "f61f78fb-c52e-4df9-8b1a-2219b2188bd4"
	connectionUUID := "894e3aa3-8beb-4e79-9785-3bd35fa671bc"                                                // Will also be the pod name, being injected via uuidConnectionFaker
	sessionUUID := []string{"ff8a5861-824b-4246-9457-94f04d665d7b", "2a598bf6-7591-45f1-979f-1fd07f97704c"} // Being injected via uuidSessionFaker

	uuidConnectionFaker := func() uuid.UUID {
		u, _ := uuid.Parse(connectionUUID)
		return u
	}
	sshConnectionRepo, err := ssh.NewConnectionRepository(nil, uuidConnectionFaker)
	if err != nil {
		t.Fatalf("cannot create ssh connections repo: %+v", err)
	}

	var sessionCounter = 0
	uuidSessionFaker := func() uuid.UUID {
		u, _ := uuid.Parse(sessionUUID[sessionCounter])
		sessionCounter++
		return u
	}
	sshSessionRepo, err := ssh.NewSessionRepository(nil, uuidSessionFaker)
	if err != nil {
		t.Fatalf("cannot create ssh connections repo: %+v", err)
	}

	injections = append(injections, []interface{}{sshConnectionRepo, sshSessionRepo}...)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s", deviceUUID), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceID": deviceUUID})

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Create second test server with the echo handler.
	deviceServer := httptest.NewServer(http.HandlerFunc(echo))
	defer deviceServer.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u = "ws" + strings.TrimPrefix(deviceServer.URL, "http")

	// Connect to the server
	deviceServerWS, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer deviceServerWS.Close()

	go func() {
		//Bring up the fake device after a few seconds
		time.Sleep(time.Second * 3)
		sshConnectionRepo.Add(&ssh.Connection{PodName: connectionUUID, Conn: deviceServerWS}, deviceUUID)
	}()

	go func() {
		//Bring up the first fake session
		time.Sleep(time.Second * 5)
		session := ssh.Session{
			ID:         sessionUUID[0],
			DeviceConn: deviceServerWS,
		}
		sshSessionRepo.Add(&session)

		//Bring up the second fake session
		time.Sleep(time.Second * 5)
		session = ssh.Session{
			ID:         sessionUUID[1],
			DeviceConn: deviceServerWS,
		}
		sshSessionRepo.Add(&session)
	}()

	svc := NewService(injections...)
	err = svc.SSHCreate(ws, req)

	var createProcess ssh.CreateProcessMessage
	err = deviceServerWS.ReadJSON(&createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID[0] {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}
	if string(createProcess.Command) != "/dbclient -y root@127.0.0.1" {
		t.Fatalf("bad message" + string(createProcess.Command))
	}

	if err = ws.WriteJSON(ssh.NewExecuteCommandMessage(sessionUUID[0], []byte("whoami"))); err != nil {
		t.Fatal("write", err)
	}

	var executeCommand ssh.ExecuteCommandMessage
	err = deviceServerWS.ReadJSON(&executeCommand)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(executeCommand.Command) != "whoami" {
		t.Fatalf("bad message" + string(executeCommand.Command))
	}

	err = svc.SSHCreate(ws, req)

	err = deviceServerWS.ReadJSON(&createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID[1] {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}
	if string(createProcess.Command) != "/dbclient -y root@127.0.0.1" {
		t.Fatalf("bad message" + string(createProcess.Command))
	}
}
