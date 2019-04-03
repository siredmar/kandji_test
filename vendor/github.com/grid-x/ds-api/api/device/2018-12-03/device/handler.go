package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/messaging"
	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/grid-x/ds-api/types"
	v20181203 "github.com/grid-x/ds-api/types/device/2018-12-03/device"
)

const (
	version = "2018-12-03"
	group   = "device"
)

var (
	// defaultTimeout is the default timeout used when calling downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	DeviceIDFromContext(context.Context) (string, error)
	AccountIDFromContext(context.Context) (string, error)
}

type deviceClient interface {
	Get(ctx context.Context, namespace, name string) (*corev1beta1.Device, error)
	UpdateStatus(ctx context.Context, dev *corev1beta1.Device) (*corev1beta1.Device, error)
}

type podClient interface {
	Create(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)
	Delete(ctx context.Context, namespace, name string) error
}

type natsRepository interface {
	// Subscribes to ssh messages coming from 1-n clients to route them to the connected device
	SubscribeToSSHMessagesForDevice(deviceID string, sshC chan []byte) (*nats.Subscription, error)
	// Publish messages coming from the connected device to a certain client
	PublishSSHMessageToClient(deviceID string, sessionID string, b []byte) error
	// Listens for ping requests from 1-n clients and respond with a Pong
	ListenForClientPing(deviceID string) (*nats.Subscription, error)
	// Listens for registration requests from 1-n clients to create a new session
	ListenForClientReg(deviceID string, sessionC chan string) (*nats.Subscription, error)
}

type uuidRepository interface {
	Get() uuid.UUID
}

type durationRepository interface {
	GetSSHInactiveTimeout() time.Duration
	GetSSHInactiveTicker() time.Duration
}

// Service implements the HTTP endpoints for the device API
type Service struct {
	logger       log.FieldLogger
	deviceClient deviceClient
	podClient    podClient
	ap           authProvider
	nats         natsRepository
	uuid         uuidRepository
	timeout      durationRepository
}

func deviceFromK8s(dev *corev1beta1.Device) *v20181203.Device {
	return &v20181203.Device{
		Metadata: types.ConvertFromK8sMetadata(dev.ObjectMeta, true),
		Spec: v20181203.DeviceSpec{
			AccountID:         dev.Spec.AccountID,
			Serialnumber:      dev.Spec.Serialnumber,
			MenderDeviceID:    dev.Spec.MenderDeviceID,
			PublicKey:         dev.Spec.PublicKey,
			MACAddress:        dev.Spec.MACAddress,
			MaintenanceWindow: dev.Spec.MaintenanceWindow,
		},
		Status: statusFromK8s(dev.Status),
	}
}

func statusFromK8s(s corev1beta1.DeviceStatus) v20181203.DeviceStatus {
	return v20181203.DeviceStatus{
		LastHeartbeat: s.LastHeartbeat,
	}
}

func statusToK8s(s v20181203.DeviceStatus) corev1beta1.DeviceStatus {
	return corev1beta1.DeviceStatus{
		LastHeartbeat: s.LastHeartbeat,
	}
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case deviceClient:
			s.deviceClient = i
		case podClient:
			s.podClient = i
		case authProvider:
			s.ap = i
		case natsRepository:
			s.nats = i
		case uuidRepository:
			s.uuid = i
		case durationRepository:
			s.timeout = i
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

// GetResponse represents the response type
type GetResponse v20181203.GetResponse

// WriteText creates a text representation from GetResponse
func (resp *GetResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from GetResponse
func (resp *GetResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Get implements the HTTP handler for getting device status and spec
//
// @name: GetDevice
// @description: Get the authenticated device
// @action: device:Get
// @resource: device
// @endpoint: GET /
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.ap.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID, err := s.ap.DeviceIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get deviceID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err := s.deviceClient.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with ID %s does not exist", deviceID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			Device: deviceFromK8s(dev),
		},
	}, nil
}

// UpdateRequest represents the request type
type UpdateRequest v20181203.UpdateRequest

// Validate validates an UpdateRequest
func (req *UpdateRequest) Validate() error {
	if req.Status == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing status"),
		)
	}

	return nil
}

// ReadJSON reads UpdateRequest from a JSON payload
func (req *UpdateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// UpdateResponse represents the response type
type UpdateResponse v20181203.UpdateResponse

// WriteText creates a text representation from UpdateResponse
func (resp *UpdateResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from UpdateResponse
func (resp *UpdateResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Update implements the HTTP handler for the update device endpoint
//
// @name: UpdateDevice
// @description: Updates the device status
// @action: device:Update
// @resource: device
// @endpoint: PATCH /
// @middlewares: auth
func (s *Service) Update(req *http.Request, payload UpdateRequest) (*encoding.Response, error) {
	accountID, err := s.ap.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID, err := s.ap.DeviceIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get deviceID: %+v", err),
		)
	}

	if payload.Status == nil {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err := s.deviceClient.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Device with ID %s does not exist", deviceID),
		)
	}
	dev.Status = statusToK8s(*payload.Status)

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	dev, err = s.deviceClient.UpdateStatus(ctx, dev)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch device %s: %+v", dev.Name, err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateResponse{
			Device: deviceFromK8s(dev),
		},
	}, nil
}

// SSHAgentConnect implements the HTTP handler for creating a ssh connection
//
// @name: SSHAgentConnect
// @description: Connects an agent
// @action: device:SSHAgentConnect
// @resource: device
// @protocol: WS
// @endpoint: GET /ssh
// @middlewares: auth
func (s *Service) SSHAgentConnect(conn *websocket.Conn, req *http.Request) {
	// Make sure we close client socket connection properly
	defer conn.Close()

	// Inialize a new WebsocketWriter to avoid concurrent writes
	socketWriter := messaging.NewWebsocketWriter(conn)

	accountID, err := s.ap.AccountIDFromContext(req.Context())
	if err != nil {
		handleWebsocketError(socketWriter, "Cannot get accountID", "Internal server error", err, s.logger)
		return
	}

	deviceID, err := s.ap.DeviceIDFromContext(req.Context())
	if err != nil {
		handleWebsocketError(socketWriter, "Cannot get deviceID", "Internal server error", err, s.logger)
		return
	}

	logger := s.logger.WithField("deviceID", deviceID)
	logger.Infof("Device connected.")

	// Check every 60sec if the pod should be delete due to inactivity
	inactiveTick := time.NewTicker(s.timeout.GetSSHInactiveTicker())
	defer inactiveTick.Stop()

	lastCommand := time.Now()

	var expired bool
	go func() {
		for {
			select {
			case <-inactiveTick.C:
				t := time.Now()
				l := &lastCommand

				expiresAt := l.Add(s.timeout.GetSSHInactiveTimeout())

				if t.After(expiresAt) {
					// SSH connection expired. Close the socket in order to get the regular cleanup process kicked in
					conn.Close()
					expired = true
				}
			}
			if expired {
				break
			}
		}
	}()

	// Start listener for device pings
	pingSub, err := s.nats.ListenForClientPing(deviceID)
	if err != nil {
		handleWebsocketError(socketWriter, "Could not subscribe to client pings", "Internal server error", err, logger)
		return
	}
	defer pingSub.Unsubscribe()

	// Start Listener for new sessions
	sessionC := make(chan string)
	defer close(sessionC)
	sessionSub, err := s.nats.ListenForClientReg(deviceID, sessionC)
	if err != nil {
		handleWebsocketError(socketWriter, "Could not subscribe to client regs", "Internal server error", err, logger)
		return
	}
	defer sessionSub.Unsubscribe()

	go func() {
		for {
			// Received new client reg (Format sessionID$initCommand)
			sessionReg, ok := <-sessionC
			if !ok {
				break
			}

			splitted := strings.Split(sessionReg, "$")
			sessionID := splitted[0]
			initCommand := splitted[1]

			logger.WithField("sessionID", sessionID).Infof("Starting new session")
			init := ssh.NewCreateProcessMessage(sessionID, []byte(initCommand))

			err := socketWriter.WriteJSON(init)
			if err != nil {
				logger.WithField("sessionID", sessionID).Errorf("Device could not init a new session! Reason: %s", err)
			}
		}
	}()

	// Subscribe to messages from clients on the device channel
	sshClientC := make(chan []byte)

	sub, err := s.nats.SubscribeToSSHMessagesForDevice(deviceID, sshClientC)
	if err != nil {
		close(sshClientC)
		handleWebsocketError(socketWriter, "Could not subscribe to ssh messages", "Internal server error", err, logger)
		return
	}
	defer sub.Unsubscribe()

	// Forward messages from device to client
	go func() {
		for {
			var messageType ssh.RAWMessage

			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Infof("Device disconnected")

				// Todo...
				// + Inform clients
				// + Store Session to DB
				break
			}

			err = json.Unmarshal(message, &messageType)
			if err != nil {
				handleWebsocketError(socketWriter, "Unmarshal error on the deviceChannel", "Internal server error", err, logger)
				break
			}

			switch messageType.Type {
			case ssh.ProcessOutputMessageType:
				var output ssh.ProcessOutputMessage
				err = json.Unmarshal(message, &output)
				if err != nil {
					handleWebsocketError(socketWriter, "Not able to unmarshall process output", "Internal server error", err, logger)
					break
				}
				// Forward to the client on session channel
				if err = s.nats.PublishSSHMessageToClient(deviceID, output.ID, message); err != nil {
					handleWebsocketError(socketWriter, "Not able to publish to ssh messages", "Internal server error", err, logger.WithField("sessionID", output.ID))
					break
				}
			case ssh.ProcessCreatedMessageType:
				var created ssh.ProcessCreatedMessage
				err = json.Unmarshal(message, &created)
				if err != nil {
					handleWebsocketError(socketWriter, "Not able to unmarshall process created", "Internal server error", err, logger)
					break
				}

				// Forward to the client on session channel
				if err = s.nats.PublishSSHMessageToClient(deviceID, created.ID, message); err != nil {
					handleWebsocketError(socketWriter, "Not able to publish to ssh messages", "Internal server error", err, logger.WithField("sessionID", created.ID))
					break
				}
			case ssh.ProcessTerminatedMessageType:
				var terminated ssh.ProcessTerminatedMessage
				err = json.Unmarshal(message, &terminated)
				if err != nil {
					handleWebsocketError(socketWriter, "Not able to unmarshall process created", "Internal server error", err, logger)
					break
				}
				// Forward to the client on session channel
				if err = s.nats.PublishSSHMessageToClient(deviceID, terminated.ID, message); err != nil {
					handleWebsocketError(socketWriter, "Not able to subscribe to ssh messages", "Internal server error", err, logger.WithField("sessionID", terminated.ID))
					break
				}
			default:
				logger.Infof("Received an unknown message type on deviceChannel: %v", string(message))
			}
		}

		// At this point we stopt communicating with the device. Close sshClientC in order to exit the client loop and return
		close(sshClientC)

		// Remove Pod
		logger.Infof("Deleting SSH pod")
		ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
		defer cancel()

		err := s.podClient.Delete(ctx, model.AccountNamespaceName(accountID), deviceID+"-ssh")
		if err != nil {
			logger.Errorf("Pod '%s' could not be deleted after being inactive: %+v", deviceID+"-ssh", err)
		}
	}()

	var messageType ssh.RAWMessage
	for {
		msg, ok := <-sshClientC
		if !ok {
			break
		}

		err = json.Unmarshal(msg, &messageType)
		if err != nil {
			handleWebsocketError(socketWriter, "Unmarshal error on the deviceChannel", "Internal server error", err, logger)
			return
		}

		lastCommand = time.Now()

		// Determine kind of message
		switch messageType.Type {
		case ssh.ExecuteCommandMessageType:
			var execute ssh.ExecuteCommandMessage
			err = json.Unmarshal(msg, &execute)
			if err != nil {
				handleWebsocketError(socketWriter, "Not able to unmarshall execute command", "Internal server error", err, logger)
				return
			}

			// Message could be correctly unmarshaled, forward it to the client
			if err = socketWriter.WriteJSON(execute); err != nil {
				handleWebsocketError(socketWriter, "Not able to forward message to the device", "Internal server error", err, logger)
				return
			}
		default:
			logger.Infof("Received an unknown message type on clientChannel: %v", string(msg))
		}
	}

	logger.Infof("SSH connection for device closed")
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
