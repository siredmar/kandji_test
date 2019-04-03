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
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/messaging"
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
	Get(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error)
}

type natsRepository interface {
	// Subscribe to messages coming from a certain device
	SubscribeToSSHMessagesForClient(deviceID string, sessionID string, sshC chan []byte) (*nats.Subscription, error)
	// Publish messages coming from the client to a certain device
	PublishSSHMessageToDevice(deviceID string, sessionID string, b []byte) error
	// Send a ping request to a device to check if it is connected
	SendPingToDevice(deviceID string, sessionID string, timeout time.Duration) error
	// Register a new client session with a device
	RegisterAtDevice(deviceID string, sessionID string, initCommand string, timeout time.Duration) error
}

type uuidRepository interface {
	Get() uuid.UUID
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
	LastHeartbeat string `json:"lastHeartbeat,omitempty"`
}

func fromK8sType(d *corev1beta1.Device) *Device {
	var lastHeartbeat string
	if d.Status.LastHeartbeat != nil {
		lastHeartbeat = d.Status.LastHeartbeat.UTC().Format(time.RFC3339)
	}

	return &Device{
		Metadata: api.ConvertFromK8sMetadata(d.ObjectMeta, false),
		Spec: DeviceSpec{
			Serialnumber:      d.Spec.Serialnumber,
			MACAddress:        d.Spec.MACAddress,
			PublicKey:         d.Spec.PublicKey,
			MaintenanceWindow: &d.Spec.MaintenanceWindow,
		},
		Status: DeviceStatus{
			LastHeartbeat: lastHeartbeat,
		},
	}
}

// Service implements the handlers for this group
type Service struct {
	logger    log.FieldLogger
	auth      authProvider
	client    deviceRepository
	podClient podRepository
	nats      natsRepository
	uuid      uuidRepository
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
		case natsRepository:
			s.nats = i
		case uuidRepository:
			s.uuid = i
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
func (s *Service) SSHCreate(conn *websocket.Conn, req *http.Request) {
	// Make sure we close client socket connection properly
	defer conn.Close()

	// Inialize a new WebsocketWriter to avoid concurrent writes
	socketWriter := messaging.NewWebsocketWriter(conn)

	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		handleWebsocketError(socketWriter, "Cannot get accountID", "Internal server error", err, s.logger)
		return
	}

	deviceID := mux.Vars(req)["deviceID"]
	if deviceID == "" {
		handleWebsocketError(socketWriter, "Missing Device ID", "Missing device id", err, s.logger)
		return
	}

	// eg. /dbclient -y root@127.0.0.1
	initCommand := req.Header.Get("command")
	if initCommand == "" {
		handleWebsocketError(socketWriter, "Missing command", "Missing command", err, s.logger)
	}

	// Generate a new sessionID for this client connection
	sessionID := s.uuid.Get().String()

	logger := s.logger.WithFields(log.Fields{"deviceID": deviceID, "sessionID": sessionID})
	logger.Infof("Client requested ssh connection for device: %s. Start session: %s", deviceID, sessionID)

	// Check if the device is already connected by sending ping to the device handler
	err = s.nats.SendPingToDevice(deviceID, sessionID, 5*time.Second)

	if err != nil {
		// Device is not connected yet as ping did not work. Eventually create a new pod and wait for the device to be connected.
		// Check if there is a pod up already, eg. by a second concurrently running client connection
		ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
		defer cancel()

		pod, _ := s.podClient.Get(ctx, model.AccountNamespaceName(accountID), deviceID+"-ssh", true)

		if pod == nil {
			logger.Infof("No ssh pod up yet, will create a new one for device: %s", deviceID)

			ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
			defer cancel()

			podtemplate := getSSHPodTemplate(deviceID, accountID)
			pod, err := s.podClient.Create(ctx, podtemplate)
			if err != nil {
				handleWebsocketError(socketWriter, "Cannot create pod for ssh connection", "Internal server error", err, logger)
				return
			}
			logger.Infof("Pod %s created to server SSH connections", pod.Name)
		}

		// Wait up to 30 secs until device connection comes up
		connectionTick := time.NewTicker(3 * time.Second)
		defer connectionTick.Stop()

		connectionTimer := time.NewTimer(30 * time.Second)
		defer connectionTimer.Stop()

		var connected bool
		for {
			select {
			case <-connectionTimer.C:
				handleWebsocketError(socketWriter, "Device connection did not came up after creating the pod for ssh connection", "Internal server error", err, logger)
				return
			case <-connectionTick.C:
				// Ping the device handler
				err = s.nats.SendPingToDevice(deviceID, sessionID, 2*time.Second)
				if err == nil {
					connected = true
				}
			}
			if connected {
				break
			}
		}
	}
	logger.Infof("Got device connection")

	// Register Session
	err = s.nats.RegisterAtDevice(deviceID, sessionID, initCommand, 3*time.Second)
	if err != nil {
		handleWebsocketError(socketWriter, "Client could not register new session", "Internal server error", err, logger)
		return
	}

	// Subscribe to messages from the device on the client channel
	sshC := make(chan []byte)

	sub, err := s.nats.SubscribeToSSHMessagesForClient(deviceID, sessionID, sshC)
	if err != nil {
		close(sshC)
		handleWebsocketError(socketWriter, "Could not subscribe to ssh messages", "Internal server error", err, logger)
		return
	}
	defer sub.Unsubscribe()

	// Forward messages from client to device
	go func() {
		for {
			var messageType ssh.RAWMessage
			_, message, err := conn.ReadMessage()

			if err != nil {
				logger.Infof("Client disconnected")
				break
			}

			err = json.Unmarshal(message, &messageType)
			if err != nil {
				handleWebsocketError(socketWriter, "Unmarshal error on the clientChannel", "Internal server error", err, logger)
				break
			}

			switch messageType.Type {
			case ssh.CreateProcessMessageType:
				// Forward to the device
				if err = s.nats.PublishSSHMessageToDevice(deviceID, sessionID, message); err != nil {
					handleWebsocketError(socketWriter, "Message could not be forwarded to the device", "Internal server error", err, logger)
					break
				}
			case ssh.ExecuteCommandMessageType:
				// Forward to the device
				if err = s.nats.PublishSSHMessageToDevice(deviceID, sessionID, message); err != nil {
					handleWebsocketError(socketWriter, "Message could not be forwarded to the device", "Internal server error", err, logger)
					break
				}
			default:
				logger.Infof("Received an unknown message on the clientChannel: %v", string(message))
			}
		}

		// At this point we stopt communicating with the client. Close sshC in order to exit the client loop and return
		close(sshC)
	}()

	var messageType ssh.RAWMessage
	for {
		msg, ok := <-sshC
		if !ok {
			break
		}

		err = json.Unmarshal(msg, &messageType)
		if err != nil {
			handleWebsocketError(socketWriter, "Unmarshal error on the deviceChannel", "Internal server error", err, logger)
			return
		}

		// Determine kind of message
		switch messageType.Type {
		case ssh.ProcessOutputMessageType:
			var output ssh.ProcessOutputMessage
			err = json.Unmarshal(msg, &output)
			if err != nil {
				handleWebsocketError(socketWriter, "Not able to unmarshall process output", "Internal server error", err, logger)
				return
			}

			// Message could be correctly unmarshaled, forward it to the client
			if err = socketWriter.WriteJSON(output); err != nil {
				handleWebsocketError(socketWriter, "Not able to forward message to the client", "Internal server error", err, logger)
				return
			}
		case ssh.ProcessCreatedMessageType:
			var created ssh.ProcessCreatedMessage
			err = json.Unmarshal(msg, &created)
			if err != nil {
				handleWebsocketError(socketWriter, "Not able to unmarshall process created", "Internal server error", err, logger)
				return
			}

			// Message could be correctly unmarshaled, forward it to the client
			if err = socketWriter.WriteJSON(created); err != nil {
				handleWebsocketError(socketWriter, "Not able to forward message to the client", "Internal server error", err, logger)
				return
			}
		case ssh.ProcessTerminatedMessageType:
			var terminated ssh.ProcessTerminatedMessage
			err = json.Unmarshal(msg, &terminated)
			if err != nil {
				handleWebsocketError(socketWriter, "Not able to unmarshall process created", "Internal server error", err, logger)
				return
			}

			// Message could be correctly unmarshaled, forward it to the client
			if err = socketWriter.WriteJSON(terminated); err != nil {
				handleWebsocketError(socketWriter, "Not able to forward message to the client", "Internal server error", err, logger)
				return
			}

			// Process has terminated, exit
			return
		default:
			logger.Infof("Received Agent an unknown message type: %v", string(msg))
		}
	}

	return
}

func handleWebsocketError(writer *messaging.WebsocketWriter, internalMsg, externalMsg string, err error, logger log.FieldLogger) {
	internal := fmt.Sprintf("%s: %+v", internalMsg, err)
	external := externalMsg

	// Add internal log
	logger.Errorf(internal)

	// Forward error to the device
	e := ssh.NewErrorMessage(external)
	writer.WriteJSON(e)
}

func getSSHPodTemplate(deviceID, accountID string) *corev1beta1.DevicePod {
	var socket corev1beta1.HostPathType = "Socket"
	return &corev1beta1.DevicePod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deviceID + "-ssh",
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
}
