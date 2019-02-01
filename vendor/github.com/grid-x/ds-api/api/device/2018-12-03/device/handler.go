package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
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

// Service implements the HTTP endpoints for the device API
type Service struct {
	logger       log.FieldLogger
	deviceClient deviceClient
	ap           authProvider
}

// Device as exported by this API version
type Device struct {
	Metadata api.Metadata             `json:"metadata"`
	Spec     corev1beta1.DeviceSpec   `json:"spec"`
	Status   corev1beta1.DeviceStatus `json:"status"`
}

func deviceFromK8s(dev *corev1beta1.Device) *Device {
	return &Device{
		Metadata: api.ConvertFromK8sMetadata(dev.ObjectMeta),
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
		case authProvider:
			s.ap = i
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
