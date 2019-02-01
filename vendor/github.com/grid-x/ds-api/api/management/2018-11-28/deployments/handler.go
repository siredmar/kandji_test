package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-11-28"
	group   = "deployments"
)

var (
	// defaultTimeout is the default timeout used when calling downstream
	// services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(context.Context) (string, error)
}

type k8sDeploymentClient interface {
	Create(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
	Update(ctx context.Context, deploy *appsv1beta1.DeviceDeployment) (*appsv1beta1.DeviceDeployment, error)
	Delete(ctx context.Context, namespace, name string) error
	Get(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceDeployment, error)
	List(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceDeployment, error)
}

type applicationRepository interface {
	Get(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceApplication, error)
}

// Deployment represents the deployment type as exported by this API version
type Deployment struct {
	Metadata api.Metadata                       `json:"metadata,omitempty"`
	Spec     appsv1beta1.DeviceDeploymentSpec   `json:"spec"`
	Status   appsv1beta1.DeviceDeploymentStatus `json:"status"`
}

func deployFromK8s(deploy *appsv1beta1.DeviceDeployment) *Deployment {
	return &Deployment{
		Metadata: api.ConvertFromK8sMetadata(deploy.ObjectMeta),
		Spec:     deploy.Spec,
		Status:   deploy.Status,
	}
}

func newK8sDeployment(accountID, id string, spec appsv1beta1.DeviceDeploymentSpec) *appsv1beta1.DeviceDeployment {
	return &appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: accountID,
			Name:      id,
		},
		Spec:   spec,
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}
}

// Service implements the HTTP handlers for the deployment endpoints of this
// version
type Service struct {
	logger log.FieldLogger

	appRepo   applicationRepository
	k8sDeploy k8sDeploymentClient

	auth authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case applicationRepository:
			s.appRepo = i
		case k8sDeploymentClient:
			s.k8sDeploy = i
		case authProvider:
			s.auth = i
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
	Spec *appsv1beta1.DeviceDeploymentSpec `json:"spec"`
}

// Validate validates a CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Spec == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing spec"),
		)
	}
	if req.Spec.App == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing application name"),
		)
	}
	if req.Spec.Selector.MatchByLabels == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing labels to match"),
		)
	}
	if req.Spec.Template.Spec.Containers == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing containers definition"),
		)
	}

	return nil
}

// ReadJSON reads CreateRequest from a JSON payload
func (req *CreateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Deployment
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

// Create implements the HTTP handler for creating deployments
//
// @name: CreateDeployment
// @description: Creates a new deployment
// @action: deployments:Create
// @resource: deployments:*
// @endpoint: POST /deployments
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

	_, err = s.appRepo.Get(ctx, model.AccountNamespaceName(accountID), payload.Spec.App)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Application with name %s does not exist", payload.Spec.App),
		)
	}

	deploy := newK8sDeployment(model.AccountNamespaceName(accountID), uuid.NewV4().String(), *payload.Spec)

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	deploy, err = s.k8sDeploy.Create(ctx, deploy)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create deployment %s: %+v", deploy.Name, err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{Deployment: deployFromK8s(deploy)},
	}, nil
}

// ListResponse represents the response type
type ListResponse struct {
	Deployments []*Deployment `json:"deployments"`
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

// List implements the HTTP handler to list deployments
//
// @name: ListDeployments
// @description: List deployments in account
// @action: deployments:List
// @resource: deployments:*
// @endpoint: GET /deployments
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

	ds, err := s.k8sDeploy.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get deployment list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Name < ds[j].Name
	})

	result := make([]*Deployment, len(ds))
	for i, d := range ds {
		result[i] = deployFromK8s(d)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Deployments: result,
		},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	*Deployment
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

// Get implements the HTTP handler to get deployments
//
// @name: GetDeployment
// @description: Gets a specific deployment
// @action: deployments:Get
// @resource: deployments:{deploymentID}
// @endpoint: GET /deployments/{deploymentID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deploymentID := mux.Vars(req)["deploymentID"]
	if deploymentID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing deployment id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	deploy, err := s.k8sDeploy.Get(ctx, model.AccountNamespaceName(accountID), deploymentID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Deployment with ID %s does not exist", deploymentID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			Deployment: deployFromK8s(deploy),
		},
	}, nil
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Spec *appsv1beta1.DeviceDeploymentSpec `json:"spec"`
}

// Validate validates an UpdateRequest
func (req *UpdateRequest) Validate() error {
	if req.Spec == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing spec"),
		)
	}
	if req.Spec.App == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing application name"),
		)
	}
	if req.Spec.Selector.MatchByLabels == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing labels to match"),
		)
	}
	if req.Spec.Template.Spec.Containers == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing containers definition"),
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
	*Deployment
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

// Update implements the HTTP handler for updating deployments
//
// @name: UpdateDeployment
// @description: Updates a deployment
// @action: deployments:Update
// @resource: deployments:{deploymentID}
// @endpoint: PATCH /deployments/{deploymentID}
// @middlewares: auth
func (s *Service) Update(req *http.Request, payload UpdateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deploymentID := mux.Vars(req)["deploymentID"]
	if deploymentID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing deployment id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.appRepo.Get(ctx, model.AccountNamespaceName(accountID), payload.Spec.App)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Application with name %s does not exist", payload.Spec.App),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	deploy, err := s.k8sDeploy.Get(ctx, model.AccountNamespaceName(accountID), deploymentID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Deployment with ID %s does not exist", deploymentID),
		)
	}
	deploy.Spec = *payload.Spec
	deploy.Namespace = model.AccountNamespaceName(accountID)

	ctx, cancel = context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()

	deploy, err = s.k8sDeploy.Update(ctx, deploy)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch deployment %s: %+v", deploymentID, err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateResponse{
			Deployment: deployFromK8s(deploy),
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

// Delete implements the HTTP handler that deletes deployments
//
// @name: DeleteDeployment
// @description: Deletes an existing deployment
// @action: deployments:Delete
// @resource: deployments:{deploymentID}
// @endpoint: DELETE /deployments/{deploymentID}
// @middlewares: auth
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	deploymentID := mux.Vars(req)["deploymentID"]
	if deploymentID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing deployment id name"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.k8sDeploy.Get(ctx, model.AccountNamespaceName(accountID), deploymentID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Deployment with ID %s does not exist", deploymentID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	if err := s.k8sDeploy.Delete(ctx, model.AccountNamespaceName(accountID), deploymentID); err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete deployment %s: %+v", deploymentID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil

}
