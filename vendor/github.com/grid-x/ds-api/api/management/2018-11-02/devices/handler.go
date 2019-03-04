package devices

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/ssh"
)

const (
	version = "2018-11-02"
	group   = "devices"
)

var (
	// defaultTimeout is the default timeout used when calling downstream
	// services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
}

type deviceRepository interface {
	Create(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error)
	Get(ctx context.Context, namespace string, id string) (*corev1beta1.Device, error)
	GetBySerialnumber(ctx context.Context, serialnumber string) (*corev1beta1.Device, error)
	List(ctx context.Context, namepsace string) ([]*corev1beta1.Device, error)
	Update(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error)
	Delete(ctx context.Context, namespace string, id string) error
}

type podRepository interface {
	Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)
}

type connectionRepository interface {
	// Get retrieves the stored websocket connection of an device. Returns nil if not yet connected
	Get(deviceID string) (connection *ssh.Connection)
	// Get retrieves a random v4 uuid
	GetUUID() uuid.UUID
}

type sessionRepository interface {
	// Get retrieves the stored ssh session. Returns nil if not yet setup
	Get(sessionID string) *ssh.Session
	// Get retrieves a random v4 uuid
	GetUUID() uuid.UUID
}

// Device as used by this API version
type Device struct {
	Metadata api.Metadata `json:"metadata"`
	Spec     DeviceSpec   `json:"spec,omitempty"`
	Status   DeviceStatus `json:"status,omitempty"`
}

// DeviceSpec represents the spec of a device
type DeviceSpec struct {
	Serialnumber      string  `json:"serialnumber"`
	MACAddress        *string `json:"macAddress,omitempty"`
	PublicKey         *string `json:"publicKey,omitempty"`
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
}

// DeviceStatus represents the status of a device
type DeviceStatus struct {
	LastHeartbeat *string `json:"lastHeartbeat,omitempty"`
}

func fromK8sType(d *corev1beta1.Device) *Device {
	return &Device{
		Metadata: api.ConvertFromK8sMetadata(d.ObjectMeta, false),
		Spec: DeviceSpec{
			Serialnumber:      d.Spec.Serialnumber,
			MACAddress:        d.Spec.MACAddress,
			PublicKey:         d.Spec.PublicKey,
			MaintenanceWindow: &d.Spec.MaintenanceWindow,
		},
		Status: DeviceStatus{
			LastHeartbeat: d.Status.LastHeartbeat,
		},
	}
}

// Service implements the handlers for this group
type Service struct {
	logger        log.FieldLogger
	auth          authProvider
	client        deviceRepository
	podClient     podRepository
	sshConnection connectionRepository
	sshSession    sessionRepository
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case authProvider:
			s.auth = i
		case deviceRepository:
			s.client = i
		case podRepository:
			s.podClient = i
		case connectionRepository:
			s.sshConnection = i
		case sessionRepository:
			s.sshSession = i
		}
	}

	if s.logger == nil {
		s.logger = log.New().WithFields(log.Fields{
			"version": version,
			"group":   group,
		})
	}

	return s
}

// CreateRequest represents the request type
type CreateRequest struct {
	Spec DeviceSpec `json:"spec"`
}

// Validate validates a CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Spec.Serialnumber == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing serialnumber"),
		)
	}
	if req.Spec.MaintenanceWindow != nil {
		_, err := model.NewMaintenanceWindow(*req.Spec.MaintenanceWindow)
		if err != nil {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Invalid maintenance window %s: %+v", *req.Spec.MaintenanceWindow, err),
			)
		}
	}

	return nil
}

// ReadJSON reads CreateRequest from a JSON payload
func (req *CreateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Device
}

// WriteText creates a text representation from CreateResponse
func (resp *CreateResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from CreateResponse
func (resp *CreateResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Create implements the handler to create a new device
//
// @name: CreateDevice
// @description: Creates a new device in the K8s cluster
// @action: devices:Create
// @resource: devices
// @endpoint: POST /devices
// @middlewares: auth
func (s *Service) Create(req *http.Request, payload CreateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	serialnumber := payload.Spec.Serialnumber
	maintenanceWindow := "Sun:04:00-Sun:06:00"
	deviceID := uuid.New()

	if payload.Spec.MaintenanceWindow != nil {
		maintenanceWindow = *payload.Spec.MaintenanceWindow
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.client.GetBySerialnumber(ctx, serialnumber)
	if err == nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with serialnumber %s already exists", serialnumber),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err := s.client.Create(ctx, &corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: model.AccountNamespaceName(accountID),
			Name:      deviceID.String(),
		},
		Spec: corev1beta1.DeviceSpec{
			AccountID:         accountID,
			MACAddress:        payload.Spec.MACAddress,
			PublicKey:         payload.Spec.PublicKey,
			Serialnumber:      serialnumber,
			MaintenanceWindow: maintenanceWindow,
		},
	})
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create device %s: %+v", serialnumber, err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{fromK8sType(dev)},
	}, nil
}

// ListResponse represents the response type
type ListResponse struct {
	Devices []*Device `json:"devices"`
}

// WriteText creates a text representation from ListResponse
func (resp *ListResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from ListResponse
func (resp *ListResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// List implements the handler to list devices
//
// @name: ListDevices
// @description: List returns the devices found in the k8s cluster
// @action: devices:List
// @resource: devices:*
// @endpoint: GET /devices
// @middlewares: auth
func (s *Service) List(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	ds, err := s.client.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get device list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Spec.Serialnumber < ds[j].Spec.Serialnumber
	})

	result := make([]*Device, len(ds))
	for i, d := range ds {
		result[i] = fromK8sType(d)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Devices: result,
		},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	*Device
}

// WriteText creates a text representation from GetResponse
func (resp *GetResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from GetResponse
func (resp *GetResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Get implements the handler to get an existing device
//
// @name: GetDevice
// @description: Get returns a device from the K8s cluster
// @action: devices:Get
// @resource: devices:{deviceID}
// @endpoint: GET /devices/{deviceID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID := mux.Vars(req)["deviceID"]
	if deviceID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing Device ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err := s.client.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with ID %s does not exist", deviceID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{fromK8sType(dev)},
	}, nil
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Metadata api.UpdateMetadata `json:"metadata"`
	Spec     UpdateSpec         `json:"spec,omitempty"`
}

// UpdateSpec represents the update spec type
type UpdateSpec struct {
	MACAddress        *string `json:"macAddress,omitempty"`
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
}

// Validate validates an UpdateRequest
func (req *UpdateRequest) Validate() error {
	if req.Spec.MACAddress == nil && req.Spec.MaintenanceWindow == nil && req.Metadata.Labels == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}
	if req.Spec.MaintenanceWindow != nil {
		_, err := model.NewMaintenanceWindow(*req.Spec.MaintenanceWindow)
		if err != nil {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Invalid maintenance window %s: %+v", *req.Spec.MaintenanceWindow, err),
			)
		}
	}

	return nil
}

// ReadJSON reads UpdateRequest from a JSON payload
func (req *UpdateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Device
}

// WriteText creates a text representation from UpdateResponse
func (resp *UpdateResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from UpdateResponse
func (resp *UpdateResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Update implements the update handler for devices
//
// @name: UpdateDevice
// @description: Update will update a device in the k8s cluster
// @action: devices:Update
// @resource: devices:{deviceID}
// @endpoint: PATCH /devices/{deviceID}
// @middlewares: auth
func (s *Service) Update(req *http.Request, payload UpdateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID := mux.Vars(req)["deviceID"]
	if deviceID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing Device ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err := s.client.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with ID %s does not exist", deviceID),
		)
	}

	if payload.Spec.MACAddress != nil {
		dev.Spec.MACAddress = payload.Spec.MACAddress
	}

	if payload.Spec.MaintenanceWindow != nil {
		dev.Spec.MaintenanceWindow = *payload.Spec.MaintenanceWindow
	}

	if payload.Metadata.Labels != nil {
		dev.Labels = model.ComputeLabels(dev.Labels, payload.Metadata.Labels)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev.Namespace = model.AccountNamespaceName(accountID)
	dev, err = s.client.Update(ctx, dev)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch device %s: %+v", dev.Name, err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateResponse{fromK8sType(dev)},
	}, nil
}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// WriteText creates a text representation from DeleteResponse
func (resp *DeleteResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from DeleteResponse
func (resp *DeleteResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Delete implements the delete handler for devices
//
// @name: DeleteDevice
// @description: Delete deletes a device in the K8s cluster
// @action: devices:Delete
// @resource: devices:{deviceID}
// @endpoint: DELETE /devices/{deviceID}
// @middlewares: auth
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID := mux.Vars(req)["deviceID"]
	if deviceID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing device ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.client.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with ID %s does not exist", deviceID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	if err := s.client.Delete(ctx, model.AccountNamespaceName(accountID), deviceID); err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete device %s: %+v", deviceID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil
}

// SSHCreate implements the HTTP handler for creating a ssh connection
//
// @name: SSHCreate
// @description: Creates a ssh connection
// @action: device:SSHCreate
// @resource: devices:{deviceID}
// @endpoint: GET /devices/{deviceID}/ssh
// @protocol: WS
// @middlewares: auth
func (s *Service) SSHCreate(conn *websocket.Conn, req *http.Request) error {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID := mux.Vars(req)["deviceID"]
	if deviceID == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing Device ID"),
		)
	}

	// Check for existing device connetions
	deviceConnection := s.sshConnection.Get(deviceID)

	if deviceConnection == nil {
		// No connection from this device yet. Create a new pod and wait for the device to be connected
		connectionToken := s.sshConnection.GetUUID().String()

		var socket corev1beta1.HostPathType = "Socket"
		sshPodTemplate := &corev1beta1.DevicePod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionToken,
				Namespace: model.AccountNamespaceName(accountID),
			},
			Spec: corev1beta1.DevicePodSpec{
				DeviceID: deviceID,
				Config: corev1beta1.PodConfig{
					Volumes: []corev1beta1.Volume{
						{
							Name: "supervisorAPI",
							VolumeSource: corev1beta1.VolumeSource{
								HostPath: &corev1beta1.HostPathVolumeSource{
									Path: "/var/run/supervisor.sock",
									Type: &socket,
								},
							},
						},
					},
					Network: "Host",
					Containers: []corev1beta1.Container{
						{
							VolumeMounts: []corev1beta1.VolumeMount{
								{
									Name:      "supervisorAPI",
									MountPath: "/var/run/supervisor.sock",
								},
							},
							Name:  "ssh-agent",
							Image: "ds-ssh-agent:latest",
							Environment: []corev1beta1.EnvVar{
								{
									Name:  "DROPBEAR_PASSWORD",
									Value: "fa",
								},
							},
						},
					},
				},
			},
		}

		ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
		defer cancel()

		pod, err := s.podClient.Create(ctx, sshPodTemplate)
		if err != nil {
			return errors.E(
				errors.Internal,
				fmt.Errorf("Cannot create pod for ssh connection %s: %+v", pod.Name, err),
			)
		}
		s.logger.Infof("Pod %s created to server SSH connections", pod.Name)

		connectionTick := time.NewTicker(50 * time.Millisecond)
		defer connectionTick.Stop()

		connectionTimer := time.NewTimer(30 * time.Second)
		defer connectionTimer.Stop()

		var connected bool
		for {
			select {
			case <-connectionTimer.C:
				return errors.E(
					errors.Internal,
					fmt.Errorf("Not able to create ssh connection"),
				)
			case <-connectionTick.C:
				deviceConnection = s.sshConnection.Get(deviceID)

				if deviceConnection != nil {
					connected = true
				}
			}
			if connected {
				break
			}
		}

		// Setting PodName into connection to delete it once the connection gets closed
		deviceConnection.PodName = pod.Name

		s.logger.Infof("Device connected")
	}

	sessionID := s.sshSession.GetUUID().String()

	// Issue initial command
	init := ssh.CreateProcessMessage{
		ID:      sessionID,
		Command: []byte("/dbclient -y root@127.0.0.1"),
	}

	s.logger.Infof("Creating session %s", sessionID)
	if err = deviceConnection.Conn.WriteJSON(init); err != nil {
		return errors.E(
			errors.Internal,
			fmt.Errorf("Not able to write message"),
		)
	}

	// Wait for session to be started
	sessionTick := time.NewTicker(50 * time.Millisecond)
	defer sessionTick.Stop()

	sessionTimer := time.NewTimer(5 * time.Second)
	defer sessionTimer.Stop()

	var session *ssh.Session
	var sessionCreated bool
	for {
		select {
		case <-sessionTimer.C:
			return errors.E(
				errors.Internal,
				fmt.Errorf("Not able to get ssh session"),
			)
		case <-sessionTick.C:
			session = s.sshSession.Get(sessionID)

			if session != nil {
				sessionCreated = true
			}
		}
		if sessionCreated {
			break
		}
	}

	session.ClientConn = conn

	go func() {
		defer conn.Close()

		for {
			var messageType ssh.MessageType

			_, message, err := conn.ReadMessage()

			if err != nil {
				s.logger.Infof("ReadMessage error on the clientChannel: %s", err)
				return
			}

			err = json.Unmarshal(message, &messageType)
			if err != nil {
				s.logger.Infof("Unmarshal error on the clientChannel: %s", err)
				return
			}

			switch messageType.Type {
			case ssh.CreateProcessMessageType:
				// Forward to the device
				if err = session.DeviceConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					return
				}
			case ssh.ExecuteCommandMessageType:
				// Forward to the device
				if err = session.DeviceConn.WriteMessage(websocket.BinaryMessage, message); err != nil {
					return
				}
			default:
				s.logger.Infof("Received an unknown message on the clientChannel: %v", message)
			}
		}
	}()

	return nil
}
