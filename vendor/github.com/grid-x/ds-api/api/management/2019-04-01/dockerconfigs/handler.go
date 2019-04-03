package dockerconfigs

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
	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2019-04-01"
	group   = "dockerconfigs"
)

var (
	// defaultTimeout is the default timeout used when calling downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
}

type dockerConfigRepository interface {
	Create(ctx context.Context, config *configv1beta1.DockerConfig) (*configv1beta1.DockerConfig, error)
	Get(ctx context.Context, namespace, name string) (*configv1beta1.DockerConfig, error)
	List(ctx context.Context, namespace string) ([]*configv1beta1.DockerConfig, error)
	Delete(ctx context.Context, namespace, name string) error
}

// DockerConfig as exposed by this API version
type DockerConfig struct {
	Metadata api.Metadata                     `json:"metadata,omitempty"`
	Spec     configv1beta1.DockerConfigSpec   `json:"spec"`
	Status   configv1beta1.DockerConfigStatus `json:"status"`
}

func dockerConfigFromK8s(config *configv1beta1.DockerConfig) *DockerConfig {
	return &DockerConfig{
		Metadata: api.ConvertFromK8sMetadata(config.ObjectMeta, false),
		Spec:     config.Spec,
		Status:   config.Status,
	}
}

func newK8sDockerConfig(accountID, id string, spec configv1beta1.DockerConfigSpec) *configv1beta1.DockerConfig {
	return &configv1beta1.DockerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: accountID,
			Name:      id,
		},
		Spec:   spec,
		Status: configv1beta1.DockerConfigStatus{},
	}
}

// Service implements the read only pod HTTP handlers for the management API
type Service struct {
	logger log.FieldLogger
	auth   authProvider
	config dockerConfigRepository
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
		case dockerConfigRepository:
			s.config = i
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
	Spec *configv1beta1.DockerConfigSpec `json:"spec"`
}

// Validate validates CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Spec == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing spec"),
		)
	}
	if req.Spec.Registry == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing registry"),
		)
	}
	if req.Spec.Credentials.AWS == nil || req.Spec.Credentials.DockerHub == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing credentials"),
		)
	}
	if req.Spec.Credentials.AWS != nil {
		if req.Spec.Credentials.AWS.AccessKey == "" {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Missing AWS accessKey"),
			)
		}
		if req.Spec.Credentials.AWS.SecretAccessKey == "" {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Missing AWS secretAccessKey"),
			)
		}
		if req.Spec.Credentials.AWS.Region == "" {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Missing AWS region"),
			)
		}
	}
	if req.Spec.Credentials.DockerHub != nil {
		if req.Spec.Credentials.DockerHub.Username == "" {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Missing DockerHub username"),
			)
		}
		if req.Spec.Credentials.DockerHub.Password == "" {
			return errors.E(
				errors.Validation,
				fmt.Errorf("Missing DockerHub password"),
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
	*DockerConfig
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

// Create implements the HTTP handler to create a dockerConfig
//
// @name: CreateDockerConfig
// @description: Creates a new dockerConfig
// @action: dockerconfigs:Create
// @resource: dockerconfigs:*
// @endpoint: POST /dockerconfigs
// @middlewares: auth
func (s *Service) Create(req *http.Request, payload CreateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	config := newK8sDockerConfig(model.AccountNamespaceName(accountID), uuid.New().String(), *payload.Spec)

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	config, err = s.config.Create(ctx, config)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create dockerConfig %+v", err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{DockerConfig: dockerConfigFromK8s(config)},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	*DockerConfig
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

// Get implements the HTTP handler that gets a single dockerConfig
//
// @name: GetDockerConfig
// @description: Gets a specific dockerConfig
// @action: dockerconfigs:Get
// @resource: dockerconfigs:{dockerConfigID}
// @endpoint: GET /dockerconfigs/{dockerConfigID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	dockerConfigID := mux.Vars(req)["dockerConfigID"]
	if dockerConfigID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing dockerConfig id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	config, err := s.config.Get(ctx, model.AccountNamespaceName(accountID), dockerConfigID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("DockerConfig with ID %s does not exist", dockerConfigID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			DockerConfig: dockerConfigFromK8s(config),
		},
	}, nil
}

// ListResponse represents the response type
type ListResponse struct {
	DockerConfigs []*DockerConfig `json:"dockerConfigs"`
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

// List implements the HTTP handler that lists all dockerConfigs of a account
//
// @name: ListDockerConfigs
// @description: List dockerConfigs in account
// @action: dockerconfigs:List
// @resource: dockerconfigs:*
// @endpoint: GET /dockerconfigs
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

	ds, err := s.config.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get dockerConfig list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Name < ds[j].Name
	})

	result := make([]*DockerConfig, len(ds))
	for i, d := range ds {
		result[i] = dockerConfigFromK8s(d)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			DockerConfigs: result,
		},
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

// Delete implements the HTTP handler that deletes dockerConfig
//
// @name: DeleteDockerConfig
// @description: Deletes an existing dockerConfig
// @action: dockerconfigs:Delete
// @resource: dockerconfigs:{dockerConfigID}
// @endpoint: DELETE /dockerconfigs/{dockerConfigID}
// @middlewares: auth
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	dockerConfigID := mux.Vars(req)["dockerConfigID"]
	if dockerConfigID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing dockerConfig id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.config.Get(ctx, model.AccountNamespaceName(accountID), dockerConfigID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("DockerConfig with ID %s does not exist", dockerConfigID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	if err := s.config.Delete(ctx, model.AccountNamespaceName(accountID), dockerConfigID); err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete dockerConfig %s: %+v", dockerConfigID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil
}
