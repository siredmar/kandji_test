package devices

import (
	"context"
	"encoding/json"
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
	"github.com/grid-x/ds-api/pkg/messaging"
	natsTest "github.com/grid-x/ds-api/pkg/messaging/test"
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
type podGetF func(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	create podCreateF
	get    podGetF
}

func (m *mockPodClient) Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return m.create(ctx, pod)
}
func (m *mockPodClient) Get(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
	return m.get(ctx, namespace, name, unfiltered)
}

type getUUIDF func() uuid.UUID

type mockUUIDClient struct {
	get getUUIDF
}

func (m *mockUUIDClient) Get() uuid.UUID {
	return m.get()
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
	now := metav1.Now().Rfc3339Copy()
	nowOutput := metav1.Now().Rfc3339Copy().UTC().Format(time.RFC3339)

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
							Status: corev1beta1.DeviceStatus{
								LastHeartbeat: &now,
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
					Device: &Device{
						Metadata: api.Metadata{
							ID: "2a598bf6-7591-45f1-979f-1fd07f97704c",
						},
						Spec: DeviceSpec{
							Serialnumber:      "serial-123",
							MaintenanceWindow: mkString("Sun:04:00-Sun:06:00"),
						},
						Status: DeviceStatus{
							LastHeartbeat: nowOutput,
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
var output = make(chan []byte)

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
		var messageType ssh.MessageType
		err = json.Unmarshal(message, &messageType)
		if err != nil {
			break
		}

		switch messageType.Type {
		case ssh.CreateProcessMessageType:
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		case ssh.ExecuteCommandMessageType:
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		case ssh.ProcessCreatedMessageType:
			output <- message
		case ssh.ProcessOutputMessageType:
			output <- message
		case ssh.ProcessTerminatedMessageType:
			output <- message
		}
	}
}

func TestSSHConcurrentConnections(t *testing.T) {
	server := natsTest.RunDefaultServer()
	defer server.Shutdown()

	nc, err := natsTest.NewEConn()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer nc.Close()

	natsRepo, err := messaging.NewNATSRepository(nc)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Some UUIDs to use within the test...
	deviceUUID := "f61f78fb-c52e-4df9-8b1a-2219b2188bd4"
	sessionUUID := "2a598bf6-7591-45f1-979f-1fd07f97704c"
	sessionUUID2 := "13d49cd8-3c17-4f73-9ccc-c898a60f842b"

	var podCreatedCounter = 0
	var podGetCounter = 0
	var uuidGetCounter = 0

	injections := []interface{}{
		&mockAuthProvider{
			f: func(ctx context.Context) (string, error) {
				return "default", nil
			},
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				if pod.Spec.Config.Network != "Host" {
					t.Fatalf("No host network in podconfig %s", pod.Spec.Config.Network)
				}
				if pod.Spec.Config.Containers[0].Image != "ds-ssh-agent:latest" {
					t.Fatalf("Wrong image name: %s", pod.Spec.Config.Containers[0].Image)
				}

				podCreatedCounter++
				return pod, nil
			},
			get: func(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
				podGetCounter++
				if podGetCounter == 1 {
					return nil, fmt.Errorf("No pod found")
				}
				return &corev1beta1.DevicePod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      deviceUUID + "-ssh",
					},
					Spec:   corev1beta1.DevicePodSpec{},
					Status: corev1beta1.DevicePodStatus{},
				}, nil
			},
		},
		&mockUUIDClient{
			get: func() uuid.UUID {
				uuidGetCounter++
				if uuidGetCounter == 1 {
					u, _ := uuid.Parse(sessionUUID)
					return u
				}
				u, _ := uuid.Parse(sessionUUID2)
				return u
			},
		},
		natsRepo,
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s", deviceUUID), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceID": deviceUUID})

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	client1, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer client1.Close()

	// Connect to the server
	client2, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer client2.Close()

	svc := NewService(injections...)
	go svc.SSHCreate(client1, req)
	go svc.SSHCreate(client2, req)

	// Wait until everything is setup
	time.Sleep(time.Second * 10)

	if podGetCounter != 2 {
		t.Fatalf("Wrong number of SSH pods checks... Wanted: %d, Got %d", 2, podGetCounter)
	}
	if podCreatedCounter != 1 {
		t.Fatalf("Wrong number of SSH pods created... Wanted: %d, Got %d", 1, podCreatedCounter)
	}

	nc.Close()
	server.Shutdown()
}

func TestSSHCreate(t *testing.T) {
	server := natsTest.RunDefaultServer()
	defer server.Shutdown()

	nc, err := natsTest.NewEConn()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer nc.Close()

	natsRepo, err := messaging.NewNATSRepository(nc)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Some UUIDs to use within the test...
	deviceUUID := "f61f78fb-c52e-4df9-8b1a-2219b2188bd4"
	sessionUUID := "2a598bf6-7591-45f1-979f-1fd07f97704c"

	injections := []interface{}{
		&mockAuthProvider{
			f: func(ctx context.Context) (string, error) {
				return "default", nil
			},
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				if pod.Spec.Config.Network != "Host" {
					t.Fatalf("No host network in podconfig %s", pod.Spec.Config.Network)
				}
				if pod.Spec.Config.Containers[0].Image != "ds-ssh-agent:latest" {
					t.Fatalf("Wrong image name: %s", pod.Spec.Config.Containers[0].Image)
				}
				return pod, nil
			},
			get: func(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
				return nil, fmt.Errorf("No pod found")
			},
		},
		&mockUUIDClient{
			get: func() uuid.UUID {
				u, _ := uuid.Parse(sessionUUID)
				return u
			},
		},
		natsRepo,
	}

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

	go func() {
		time.Sleep(time.Second * 5)

		// Bring up the fake device after a few seconds by responding to the ping request
		pingSub, err := natsRepo.ListenForClientPing(deviceUUID)
		if err != nil {
			t.Fatalf("%v", err)
		}
		defer pingSub.Unsubscribe()

		// Wait for client regs
		sessionC := make(chan string)
		defer close(sessionC)
		sessionSub, err := natsRepo.ListenForClientReg(deviceUUID, sessionC)
		if err != nil {
			t.Fatalf("%v", err)
		}
		defer sessionSub.Unsubscribe()

		sessionReg := <-sessionC
		splitted := strings.Split(sessionReg, "$")
		sessionID := splitted[0]

		if sessionID != sessionUUID {
			t.Fatalf("%v", err)
		}
	}()

	// Subscribe to messages from client
	sshDeviceC := make(chan []byte)
	defer close(sshDeviceC)
	msgSub, err := natsRepo.SubscribeToSSHMessagesForDevice(deviceUUID, sshDeviceC)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer msgSub.Unsubscribe()

	svc := NewService(injections...)
	go svc.SSHCreate(ws, req)

	// Wait until everything is setup (ping / request)
	time.Sleep(time.Second * 10)

	// Send some messages on the client channel and test if they get picked up by the client
	msg := ssh.NewProcessCreatedMessage(sessionUUID)
	marshalled, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	natsRepo.PublishSSHMessageToClient(deviceUUID, sessionUUID, marshalled)

	// Message should come up at the echo server
	var processCreated ssh.ProcessCreatedMessage
	err = json.Unmarshal(<-output, &processCreated)

	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(processCreated.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(processCreated.ID))
	}

	// Issue a command on the client side and test if it gets into the NATS subject for the device
	execute := ssh.NewExecuteCommandMessage(sessionUUID, []byte("whoami"))
	err = ws.WriteJSON(execute)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct command should get picked up by NATS subscriber for device
	var executeNew ssh.ExecuteCommandMessage
	err = json.Unmarshal(<-sshDeviceC, &executeNew)

	if !cmp.Equal(execute, executeNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(execute, executeNew))
	}

	nc.Close()
	server.Shutdown()
}
