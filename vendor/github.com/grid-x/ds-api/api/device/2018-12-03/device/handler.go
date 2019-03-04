package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/ssh"
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
	Delete(ctx context.Context, namespace, name string) error
}

type connectionRepository interface {
	// Add stores the websocket connection of an device and the corresponding pod name once the device connects to the API
	Add(connection *ssh.Connection, deviceID string)
	// Remove deletes a websocket connection of an device from the storage once all corresponding sessions are closed
	Remove(deviceID string)
	// Get retrieves the stored websocket connection of an device. Returns nil if not yet connected
	Get(deviceID string) (conn *ssh.Connection)
}

type sessionRepository interface {
	// Add stores the ssh session bridging a device connection with a client connection
	Add(session *ssh.Session)
	// Remove deletes a session once the communication between device and client is closed
	Remove(sessionID string)
	// Get retrieves the stored ssh session. Returns nil if not yet setup
	Get(sessionID string) *ssh.Session
	// GetByConnection retrieves an array of stored ssh sessions by using a connection to eg. clean the sessions once a connection breaks up.
	// Returns an empty array if not session found
	GetByConnection(conn *websocket.Conn) []*ssh.Session
}

// Service implements the HTTP endpoints for the device API
type Service struct {
	logger        log.FieldLogger
	deviceClient  deviceClient
	podClient     podClient
	ap            authProvider
	sshConnection connectionRepository
	sshSession    sessionRepository
}

// Device as exported by this API version
type Device struct {
	Metadata api.Metadata             `json:"metadata"`
	Spec     corev1beta1.DeviceSpec   `json:"spec"`
	Status   corev1beta1.DeviceStatus `json:"status"`
}

func deviceFromK8s(dev *corev1beta1.Device) *Device {
	return &Device{
		Metadata: api.ConvertFromK8sMetadata(dev.ObjectMeta, true),
		Spec:     dev.Spec,
		Status:   dev.Status,
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
type UpdateRequest struct {
	Status *corev1beta1.DeviceStatus `json:"status"`
}

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
	dev.Status = *payload.Status

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
func (s *Service) SSHAgentConnect(conn *websocket.Conn, req *http.Request) error {
	accountID, err := s.ap.AccountIDFromContext(req.Context())
	if err != nil {
		return errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deviceID, err := s.ap.DeviceIDFromContext(req.Context())
	if err != nil {
		return errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get deviceID: %+v", err),
		)
	}

	s.logger.Infof("Device '%s' connected.", deviceID)

	sshConnection := &ssh.Connection{Conn: conn}
	s.sshConnection.Add(sshConnection, deviceID)

	go func() {
		defer conn.Close()

		for {
			var messageType ssh.MessageType

			_, message, err := conn.ReadMessage()
			if err != nil {
				s.logger.Infof("Device '%s' disconnected unexpected. Clearing sessions and removing pod! Reason: %s", deviceID, err)

				ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
				defer cancel()

				err := s.podClient.Delete(ctx, model.AccountNamespaceName(accountID), sshConnection.PodName)
				if err != nil {
					s.logger.Errorf("Could not delete ssh pod %s", sshConnection.PodName)
					return
				}
				s.logger.Infof("SSH Pod %s deleted", sshConnection.PodName)

				sessions := s.sshSession.GetByConnection(conn)
				for _, ses := range sessions {
					s.logger.Infof("Session '%s' was lost due to agent failure", ses.ID)
					a := ssh.ProcessTerminatedMessage{
						ID:     ses.ID,
						Reason: []byte("Agent failure"),
					}
					if ses.ClientConn != nil {
						if err = ses.ClientConn.WriteJSON(a); err != nil {
							return
						}
					}
					s.sshSession.Remove(ses.ID)
				}
				s.sshConnection.Remove(deviceID)
				return
			}

			err = json.Unmarshal(message, &messageType)
			if err != nil {
				s.logger.Infof("Unmarshal error on the agentChannel: %s", err)
				return
			}

			switch messageType.Type {
			case ssh.ProcessOutputMessageType:
				var output ssh.ProcessOutputMessage
				err = json.Unmarshal(message, &output)
				if err != nil {
					s.logger.Errorf("Not able to unmarshall process output: %s", err)
				}

				session := s.sshSession.Get(output.ID)
				if session == nil {
					s.logger.Warnf("That's bad session was lost %s", output.ID)
				} else {
					if session.ClientConn == nil {
						// client comnnection is not yet bound, wait for it
						tick := time.NewTicker(50 * time.Millisecond)
						defer tick.Stop()

						timer := time.NewTimer(5 * time.Second)
						defer timer.Stop()

						var sessionBound bool
						for {
							select {
							case <-timer.C:
								s.logger.Errorf("Could not get client connection")
								return
							case <-tick.C:
								if session.ClientConn != nil {
									sessionBound = true
								}
							}
							if sessionBound {
								break
							}
						}
					}
					// Forward the message to the client
					if err = session.ClientConn.WriteJSON(output); err != nil {
						return
					}
				}

			case ssh.ProcessCreatedMessageType:
				var created ssh.ProcessCreatedMessage
				err = json.Unmarshal(message, &created)
				if err != nil {
					s.logger.Errorf("Not able to unmarshall process created: %s", err)
				}

				s.logger.Infof("Device '%s' created a new ssh session", deviceID)
				session := ssh.Session{
					ID:         created.ID,
					DeviceConn: conn,
				}
				s.sshSession.Add(&session)

			case ssh.ProcessTerminatedMessageType:
				var terminated ssh.ProcessTerminatedMessage
				err = json.Unmarshal(message, &terminated)
				if err != nil {
					s.logger.Errorf("Not able to unmarshall process created: %s", err)
				}

				ses := s.sshSession.Get(terminated.ID)
				if ses == nil {
					s.logger.Warnf("That's bad session was lost %s", terminated.ID)
				} else {
					if ses.ClientConn != nil {
						// Forward the message to the client
						if err = ses.ClientConn.WriteJSON(terminated); err != nil {
							return
						}
					}
					s.sshSession.Remove(ses.ID)
				}

			default:
				s.logger.Infof("Received Agent an unknown message type: %v", message)
			}
		}
	}()

	return nil
}
