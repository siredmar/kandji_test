package devices

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
	testutils "github.com/grid-x/ds-api/pkg/testing"
)

func newReqWithDeviceID(id string) *http.Request {
	req, _ := http.NewRequest("GET", "/devices/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"deviceID": id})
	return req
}

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
