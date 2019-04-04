package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/pkg/auth/device"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	v20181102 "github.com/grid-x/ds-api/types/device/2018-11-02/auth"
)

const (
	version = "2018-11-02"
	group   = "auth"

	// the header that contains the signature
	signatureHeader = "X-GRIDX"
)

type deviceRepository interface {
	// GetByPublicKey returns the device with the given public key or an
	// error if not found
	GetByPublicKey(ctx context.Context, publicKey string) (*corev1beta1.Device, error)
}

// Service implements the auth group handlers
type Service struct {
	logger log.FieldLogger

	repo deviceRepository

	jwtGen        *device.JWTGenerator
	jwtExpTimeout time.Duration
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{
		jwtExpTimeout: 7 * 24 * time.Hour,
	}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case deviceRepository:
			s.repo = i
		case *device.JWTGenerator:
			s.jwtGen = i
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

// GetTokenRequest represents the request type
type GetTokenRequest v20181102.GetTokenRequest

// Validate validates the GetTokenRequest
func (req *GetTokenRequest) Validate() error {
	if req.PublicKey == "" {
		return errors.E(
			errors.Validation,
			errors.Action("device/auth/GetToken"),
			"Missing public key",
			fmt.Errorf("Missing public key"),
		)
	}
	if req.IDData == "" {
		return errors.E(
			errors.Validation,
			errors.Action("device/auth/GetToken"),
			"Missing id data",
			fmt.Errorf("Missing id data"),
		)
	}
	return nil
}

// ReadJSON reads GetTokenRequest from a JSON payload
func (req *GetTokenRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// GetTokenResponse represents the response type
type GetTokenResponse v20181102.GetTokenResponse

// WriteText creates a text representation from GetTokenResponse
func (resp *GetTokenResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from GetTokenResponse
func (resp *GetTokenResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// GetToken implements the HTTP handler that allows devices to request tokens
//
// @name: AuthGetToken
// @description: Allows a device to request a token
// @action: auth:GetToken
// @resource: auth:token
// @endpoint: POST /auth
func (s *Service) GetToken(req *http.Request, payload GetTokenRequest) (*encoding.Response, error) {
	signature := req.Header.Get(signatureHeader)
	if len(signature) == 0 {
		return nil, errors.E(
			errors.Validation,
			errors.Action("device/auth/GetToken"),
			fmt.Sprintf("Missing %s header", signatureHeader),
			fmt.Errorf("empty %s header", signatureHeader),
		)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			errors.Action("device/auth/GetToken"),
			"Internal error",
			fmt.Errorf("Could not marshall get token request: %+v", err),
		)
	}

	// verify signature
	if err := device.Verify(signature, payload.PublicKey, data); err != nil {
		return nil, errors.E(
			errors.Permission,
			errors.Action("device/auth/GetToken"),
			"Signature verification failed",
			fmt.Errorf("Verify failed: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()

	dev, err := s.repo.GetByPublicKey(ctx, payload.PublicKey)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			errors.Action("device/auth/GetToken"),
			"Internal error",
			fmt.Errorf("Error while looking up device by public key: %+v", err),
		)
	}

	token, err := s.jwtGen.GenerateToken(
		time.Now().Add(s.jwtExpTimeout),
		dev.Name,
		dev.Spec.AccountID,
	)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			errors.Action("device/auth/GetToken"),
			"Internal error",
			fmt.Errorf("Error while generating jwt token: %+v", err),
		)
	}

	return &encoding.Response{
		Payload: &GetTokenResponse{Token: token},
	}, nil
}
