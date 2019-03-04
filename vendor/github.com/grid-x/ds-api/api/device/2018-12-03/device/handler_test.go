package device

import (
	"context"
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

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/ssh"
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

type mockPodClient struct {
	delete podDeleteF
}

func (m *mockPodClient) Delete(ctx context.Context, namespace, name string) error {
	return m.delete(ctx, namespace, name)
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

func TestSSHAgentConnect(t *testing.T) {
	injections := []interface{}{
		&mockAuthProvider{
			accountID: "default",
			deviceID:  "f61f78fb-c52e-4df9-8b1a-2219b2188bd4",
		},
		&mockPodClient{
			delete: func(ctx context.Context, namespace, name string) error {
				if name != "894e3aa3-8beb-4e79-9785-3bd35fa671bc" {
					t.Fatalf("Wrong pod got deleted")
				}
				return nil
			},
		},
	}

	sshConnectionRepo, err := ssh.NewConnectionRepository(nil, nil)
	if err != nil {
		t.Fatalf("cannot create ssh connections repo: %+v", err)
	}

	sshSessionRepo, err := ssh.NewSessionRepository(nil, nil)
	if err != nil {
		t.Fatalf("cannot create ssh connections repo: %+v", err)
	}

	injections = append(injections, []interface{}{sshConnectionRepo, sshSessionRepo}...)

	req, _ := http.NewRequest("GET", "/devices/f61f78fb-c52e-4df9-8b1a-2219b2188bd4", nil)
	req = mux.SetURLVars(req, map[string]string{"deviceID": "f61f78fb-c52e-4df9-8b1a-2219b2188bd4"})

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
	agentServer := httptest.NewServer(http.HandlerFunc(echo))
	defer agentServer.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u = "ws" + strings.TrimPrefix(agentServer.URL, "http")

	// Connect to the server
	agentServerWS, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer agentServerWS.Close()

	svc := NewService(injections...)
	err = svc.SSHAgentConnect(ws, req)

	// Check device connection
	connection := getConnection("f61f78fb-c52e-4df9-8b1a-2219b2188bd4", sshConnectionRepo)
	if connection == nil {
		t.Fatalf("Connection did not come up")
	}

	// Fake the PodName
	connection.PodName = "894e3aa3-8beb-4e79-9785-3bd35fa671bc"

	if err = ws.WriteJSON(ssh.NewProcessCreatedMessage("ff8a5861-824b-4246-9457-94f04d665d7")); err != nil {
		t.Fatal("write", err)
	}

	// Check session
	session := getSession("ff8a5861-824b-4246-9457-94f04d665d7", sshSessionRepo)
	if session == nil {
		t.Fatalf("Session did not come up")
	}

	// Connect agent...
	session.ClientConn = agentServerWS

	if err = ws.WriteJSON(ssh.NewProcessOutputMessage(session.ID, []byte("Command output"))); err != nil {
		t.Fatal("write", err)
	}

	var processOutput ssh.ProcessOutputMessage
	err = agentServerWS.ReadJSON(&processOutput)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(processOutput.Data) != "Command output" {
		t.Fatalf("bad message" + string(processOutput.Data))
	}

	if err = ws.WriteJSON(ssh.NewProcessTerminatedMessage(session.ID, []byte("Command exited"))); err != nil {
		t.Fatal("write", err)
	}

	var processTerminated ssh.ProcessTerminatedMessage
	err = agentServerWS.ReadJSON(&processTerminated)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(processTerminated.Reason) != "Command exited" {
		t.Fatalf("bad message" + string(processTerminated.Reason))
	}

	// Check session closure
	if getSession(session.ID, sshSessionRepo) != nil {
		t.Fatalf("Session did not close properly")
	}

	if err = ws.WriteJSON(ssh.NewProcessCreatedMessage("ff8a5861-824b-4246-9457-94f04d665d7")); err != nil {
		t.Fatal("write", err)
	}

	session1 := getSession("ff8a5861-824b-4246-9457-94f04d665d7", sshSessionRepo)
	if session1 == nil {
		t.Fatalf("Session1 did not come up")
	}

	if err = ws.WriteJSON(ssh.NewProcessCreatedMessage("04449e31-5818-40bc-8537-cacf7a4e9864")); err != nil {
		t.Fatal("write", err)
	}

	session2 := getSession("04449e31-5818-40bc-8537-cacf7a4e9864", sshSessionRepo)
	if session2 == nil {
		t.Fatalf("Session2 did not come up")
	}

	// Close device connection in order to check if the two open sessions are getting remove
	ws.Close()

	// Check connection closure by client
	if getConnection("f61f78fb-c52e-4df9-8b1a-2219b2188bd4", sshConnectionRepo) != nil {
		t.Fatalf("Connection did not close properly")
	}
	// Check session closure
	if getSession(session1.ID, sshSessionRepo) != nil {
		t.Fatalf("Session did not close properly")
	}
	// Check session closure
	if getSession(session2.ID, sshSessionRepo) != nil {
		t.Fatalf("Session did not close properly")
	}
}

func getSession(sessionID string, repo *ssh.SessionRepository) *ssh.Session {
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil
		case <-tick.C:
			session := repo.Get(sessionID)
			if session != nil {
				return session
			}
		}
	}
}

func getConnection(deviceID string, repo *ssh.ConnectionRepository) *ssh.Connection {
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil
		case <-tick.C:
			connection := repo.Get(deviceID)
			if connection != nil {
				return connection
			}
		}
	}
}
