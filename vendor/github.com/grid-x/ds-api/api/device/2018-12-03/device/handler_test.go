package device

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
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/messaging"
	natsTest "github.com/grid-x/ds-api/pkg/messaging/test"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/grid-x/ds-api/types"
	v20181203 "github.com/grid-x/ds-api/types/device/2018-12-03/device"
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

type podDeleteF func(ctx context.Context, namespace, name string) error
type podCreateF func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	create podCreateF
	delete podDeleteF
}

func (m *mockPodClient) Delete(ctx context.Context, namespace, name string) error {
	return m.delete(ctx, namespace, name)
}

func (m *mockPodClient) Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return m.create(ctx, pod)
}

type timeoutGetSSHInactiveTimeout func() time.Duration
type timeoutGetSSHInactiveTicker func() time.Duration

type mockTimeoutClient struct {
	getSSHInactiveTimeout timeoutGetSSHInactiveTimeout
	getSSHInactiveTicker  timeoutGetSSHInactiveTicker
}

func (m *mockTimeoutClient) GetSSHInactiveTimeout() time.Duration {
	return m.getSSHInactiveTimeout()
}

func (m *mockTimeoutClient) GetSSHInactiveTicker() time.Duration {
	return m.getSSHInactiveTicker()
}

func Test_Get(t *testing.T) {
	now := metav1.Now().Rfc3339Copy()

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
					Device: &v20181203.Device{
						Metadata: types.Metadata{
							ID: "foo",
						},
						Spec: v20181203.DeviceSpec{
							Serialnumber: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
						},
						Status: v20181203.DeviceStatus{
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
	now := metav1.NewTime(n)
	now2m := metav1.NewTime(n.Add(-2 * time.Minute))

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
				Status: &v20181203.DeviceStatus{
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
				Status: &v20181203.DeviceStatus{
					LastHeartbeat: &now,
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Device: &v20181203.Device{
						Metadata: types.Metadata{
							ID: "foo",
						},
						Spec: v20181203.DeviceSpec{
							Serialnumber: "ff8a5861-824b-4246-9457-94f04d665d7b",
						},
						Status: v20181203.DeviceStatus{
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
		if err := json.Unmarshal(message, &messageType); err != nil {
			break
		}

		switch messageType.Type {
		case ssh.ProcessOutputMessageType:
			if err := c.WriteMessage(mt, message); err != nil {
				break
			}
		case ssh.CreateFileMessageType:
			if err := c.WriteMessage(mt, message); err != nil {
				break
			}
		case ssh.WriteToFileMessageType:
			if err := c.WriteMessage(mt, message); err != nil {
				break
			}
		case ssh.CreateProcessMessageType:
			output <- message
		case ssh.ExecuteCommandMessageType:
			output <- message
		}
	}
}

func TestSSHInactiveTimeout(t *testing.T) {
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

	var podDeleted = false

	injections := []interface{}{
		&mockAuthProvider{
			accountID: "default",
			deviceID:  deviceUUID,
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				return nil, fmt.Errorf("not yet implemented")
			},
			delete: func(ctx context.Context, namespace, name string) error {
				podDeleted = true
				return nil
			},
		},
		natsRepo,
		&mockTimeoutClient{
			getSSHInactiveTimeout: func() time.Duration {
				return 10 * time.Second
			},
			getSSHInactiveTicker: func() time.Duration {
				return 1 * time.Second
			},
		},
	}

	req, _ := http.NewRequest("GET", "/devices/"+deviceUUID, nil)
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
		time.Sleep(time.Second * 3)

		// Simulate a ping request by a client
		err := natsRepo.SendPingToDevice(deviceUUID, sessionUUID, 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * 1)

		// Simulate a register request by a client
		err = natsRepo.RegisterAtDevice(deviceUUID, sessionUUID, "bash", 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	svc := NewService(injections...)
	go svc.SSHAgentConnect(ws, req)

	// Device should get a CreateProcessMessage after client has registered
	var createProcess ssh.CreateProcessMessage
	err = json.Unmarshal(<-output, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}

	time.Sleep(time.Second * 5)

	// Fake a new Execute Command message which should get picked up and send to the device
	msg := ssh.NewExecuteCommandMessage(sessionUUID, []byte("whoami"))
	marshalled, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}
	natsRepo.PublishSSHMessageToDevice(deviceUUID, sessionUUID, marshalled)

	// Should get picked up by echo server
	var execute ssh.ExecuteCommandMessage
	err = json.Unmarshal(<-output, &execute)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(execute.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(execute.ID))
	}

	time.Sleep(time.Second * 5)

	// Fake another Execute Command message which should get picked up and send to the device
	msg = ssh.NewExecuteCommandMessage(sessionUUID, []byte("whoami"))
	marshalled, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}
	natsRepo.PublishSSHMessageToDevice(deviceUUID, sessionUUID, marshalled)

	// Should get picked up by echo server
	err = json.Unmarshal(<-output, &execute)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(execute.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(execute.ID))
	}

	time.Sleep(time.Second * 11)
	// Check that pod got deleted

	if !podDeleted {
		t.Fatalf("SSH pod got not deleted after running in timeout")
	}

	nc.Close()
	server.Shutdown()
}

func TestSSHMultiDeviceConnections(t *testing.T) {
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

	injections := []interface{}{
		&mockAuthProvider{
			accountID: "default",
			deviceID:  deviceUUID,
		},
		natsRepo,
		&mockTimeoutClient{
			getSSHInactiveTimeout: func() time.Duration {
				return 1 * time.Hour
			},
			getSSHInactiveTicker: func() time.Duration {
				return 60 * time.Second
			},
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				return nil, fmt.Errorf("not yet implemented")
			},
			delete: func(ctx context.Context, namespace, name string) error {
				return nil
			},
		},
	}

	req, _ := http.NewRequest("GET", "/devices/"+deviceUUID, nil)
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
		time.Sleep(time.Second * 3)

		// Simulate a ping request by a client
		err := natsRepo.SendPingToDevice(deviceUUID, sessionUUID, 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * 1)

		// Simulate a register request by a client
		err = natsRepo.RegisterAtDevice(deviceUUID, sessionUUID, "bash", 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * 1)

		// Simulate a ping request by a client
		err = natsRepo.SendPingToDevice(deviceUUID, sessionUUID2, 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(time.Second * 1)

		// Simulate a register request by a client
		err = natsRepo.RegisterAtDevice(deviceUUID, sessionUUID2, "bash", 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	svc := NewService(injections...)
	go svc.SSHAgentConnect(ws, req)

	// Device should get a CreateProcessMessage after client has registered
	var createProcess ssh.CreateProcessMessage
	err = json.Unmarshal(<-output, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}

	// Device should get a second CreateProcessMessage after client2 has registered
	err = json.Unmarshal(<-output, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}

	time.Sleep(time.Second * 3)

	nc.Close()
	server.Shutdown()
}

func TestSSHAgentConnect(t *testing.T) {
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

	injections := []interface{}{
		&mockAuthProvider{
			accountID: "default",
			deviceID:  deviceUUID,
		},
		natsRepo,
		&mockTimeoutClient{
			getSSHInactiveTimeout: func() time.Duration {
				return 1 * time.Hour
			},
			getSSHInactiveTicker: func() time.Duration {
				return 60 * time.Second
			},
		},
		&mockPodClient{
			create: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
				return nil, fmt.Errorf("not yet implemented")
			},
			delete: func(ctx context.Context, namespace, name string) error {
				return nil
			},
		},
	}

	req, _ := http.NewRequest("GET", "/devices/"+deviceUUID, nil)
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

	// Inialize a new WebsocketWriter to avoid concurrent writes
	socketWriter := messaging.NewWebsocketWriter(ws)

	go func() {
		time.Sleep(time.Second * 5)

		// Simulate a ping request by a client
		err := natsRepo.SendPingToDevice(deviceUUID, sessionUUID, 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		// Simulate a register request by a client
		err = natsRepo.RegisterAtDevice(deviceUUID, sessionUUID, "bash", 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}

		time.Sleep(5 * time.Second)

		// Simulate a register request by a client
		err = natsRepo.RegisterAtDevice(deviceUUID, sessionUUID2, "bash", 3*time.Second)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	// Subscribe to messages from device on the sessions channel
	sshClientC := make(chan []byte)
	defer close(sshClientC)
	msgSub, err := natsRepo.SubscribeToSSHMessagesForClient(deviceUUID, sessionUUID, sshClientC)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer msgSub.Unsubscribe()

	sshClientC2 := make(chan []byte)
	defer close(sshClientC2)
	msgSub2, err := natsRepo.SubscribeToSSHMessagesForClient(deviceUUID, sessionUUID2, sshClientC2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer msgSub2.Unsubscribe()

	svc := NewService(injections...)
	go svc.SSHAgentConnect(ws, req)

	// Device should get a CreateProcessMessage after client has registered
	var createProcess ssh.CreateProcessMessage
	err = json.Unmarshal(<-output, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}

	// Fake a new Execute Command message which should get picked up and send to the device
	msg := ssh.NewExecuteCommandMessage(sessionUUID, []byte("whoami"))
	marshalled, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}
	natsRepo.PublishSSHMessageToDevice(deviceUUID, sessionUUID, marshalled)

	// Should get picked up by echo server
	var execute ssh.ExecuteCommandMessage
	err = json.Unmarshal(<-output, &execute)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(execute.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(execute.ID))
	}

	// Fake a device output for the executecommand message before and check if it's getting set into NATS subject
	outputMessage := ssh.NewProcessOutputMessage(sessionUUID, []byte("The great P."))

	err = socketWriter.WriteJSON(outputMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct output should get picked up by NATS subscriber for client
	var outputNew ssh.ProcessOutputMessage
	err = json.Unmarshal(<-sshClientC, &outputNew)

	if !cmp.Equal(outputMessage, outputNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(outputMessage, outputNew))
	}

	// Device should get a second CreateProcessMessage after another client has registered
	err = json.Unmarshal(<-output, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(createProcess.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(createProcess.ID))
	}

	// Fake a new Execute Command message for client1 which should get picked up and send to the device
	msg = ssh.NewExecuteCommandMessage(sessionUUID, []byte("echo client1"))
	marshalled, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}
	natsRepo.PublishSSHMessageToDevice(deviceUUID, sessionUUID, marshalled)

	// Should get picked up by echo server
	err = json.Unmarshal(<-output, &execute)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(execute.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(execute.ID))
	}

	// Fake a new Execute Command message for client2 which should get picked up and send to the device
	msg = ssh.NewExecuteCommandMessage(sessionUUID2, []byte("echo client2"))
	marshalled, err = json.Marshal(msg)
	if err != nil {
		t.Fatalf("%v", err)
	}
	natsRepo.PublishSSHMessageToDevice(deviceUUID, sessionUUID2, marshalled)

	// Should get picked up by echo server
	err = json.Unmarshal(<-output, &execute)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(execute.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(execute.ID))
	}

	// Fake a device output for client1 for the executecommand message before and check if it's getting set into NATS subject
	outputMessage = ssh.NewProcessOutputMessage(sessionUUID, []byte("client1"))

	err = socketWriter.WriteJSON(outputMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct output should get picked up by NATS subscriber for client1
	err = json.Unmarshal(<-sshClientC, &outputNew)

	if string(outputNew.ID) != sessionUUID {
		t.Fatalf("bad session id" + string(outputNew.ID))
	}
	if !cmp.Equal(outputMessage, outputNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(outputMessage, outputNew))
	}

	// Fake a device output for client2 for the executecommand message before and check if it's getting set into NATS subject
	outputMessage = ssh.NewProcessOutputMessage(sessionUUID2, []byte("client2"))

	err = socketWriter.WriteJSON(outputMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct output should get picked up by NATS subscriber for client2
	err = json.Unmarshal(<-sshClientC2, &outputNew)

	if string(outputNew.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(outputNew.ID))
	}
	if !cmp.Equal(outputMessage, outputNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(outputMessage, outputNew))
	}

	// Fake a device create file output for client2 and check if it's getting set into NATS subject
	createFileMessage := ssh.NewCreateFileMessage(sessionUUID2, "text.txt", false)

	err = socketWriter.WriteJSON(createFileMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct output should get picked up by NATS subscriber for client2
	var createFileMessageNew ssh.CreateFileMessage
	err = json.Unmarshal(<-sshClientC2, &createFileMessageNew)

	if string(createFileMessageNew.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(createFileMessageNew.ID))
	}
	if !cmp.Equal(createFileMessage, createFileMessageNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(createFileMessage, createFileMessageNew))
	}

	// Fake a device create file output for client2 and check if it's getting set into NATS subject
	writeFileMessage := ssh.NewWriteToFileMessage(sessionUUID2, "text.txt", []byte("abcd"), false)

	err = socketWriter.WriteJSON(writeFileMessage)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Correct output should get picked up by NATS subscriber for client2
	var writeFileMessageNew ssh.WriteToFileMessage
	err = json.Unmarshal(<-sshClientC2, &writeFileMessageNew)

	if string(writeFileMessageNew.ID) != sessionUUID2 {
		t.Fatalf("bad session id" + string(writeFileMessageNew.ID))
	}
	if !cmp.Equal(writeFileMessage, writeFileMessageNew) {
		t.Errorf("unexpected response: %s", cmp.Diff(writeFileMessage, writeFileMessageNew))
	}

	ws.Close()

	time.Sleep(5 * time.Second)

	// Simulate a ping request by a client
	err = natsRepo.SendPingToDevice(deviceUUID, sessionUUID, 3*time.Second)
	if err == nil {
		t.Fatalf("Expected an error to occur...")
	}

	nc.Close()
	server.Shutdown()
}
