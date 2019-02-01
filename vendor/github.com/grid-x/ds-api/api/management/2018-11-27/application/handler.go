package application

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
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-11-27"
	group   = "application"
)

var (
	// defaultTimeout is the default timeout used when calling downstream
	// services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(context.Context) (string, error)
}

type applicationRepository interface {
	Create(ctx context.Context, application *appsv1beta1.DeviceApplication) (*appsv1beta1.DeviceApplication, error)
	Get(ctx context.Context, namespace, name string) (*appsv1beta1.DeviceApplication, error)
	List(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceApplication, error)
	Delete(ctx context.Context, namespace, name string) error
}

type deploymentRepository interface {
	List(ctx context.Context, namespace string) ([]*appsv1beta1.DeviceDeployment, error)
}

// Application represents the application as exported by this API version
type Application struct {
	Metadata api.Metadata `json:"metadata"`
	Name     string       `json:"name"`
}

func fromK8s(app *appsv1beta1.DeviceApplication) Application {
	return Application{
		Metadata: api.ConvertFromK8sMetadata(app.ObjectMeta),
		Name:     app.Name,
	}
}

// Service implements the endpoints for applications create, list, and delete
type Service struct {
	logger log.FieldLogger

	deploymentRepo deploymentRepository
	repo           applicationRepository
	auth           authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case deploymentRepository:
			s.deploymentRepo = i
		case applicationRepository:
			s.repo = i
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
	Name string `json:"name"`
}

// Validate validates CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Name == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing application name"),
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
	Application
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

// Create implements the HTTP handler to create a new application
//
// @name: CreateApplication
// @description: Allows to create a new Application
// @action: application:create
// @resource: application:
// @endpoint: POST /applications
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

	_, err = s.repo.Get(ctx, model.AccountNamespaceName(accountID), payload.Name)
	if err == nil {
		return nil, errors.E(
			errors.Exist,
			fmt.Errorf("Application with name %s already exists", payload.Name),
		)
	}

	app := &appsv1beta1.DeviceApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      payload.Name,
			Namespace: model.AccountNamespaceName(accountID),
		},
		Spec:   appsv1beta1.DeviceApplicationSpec{},
		Status: appsv1beta1.DeviceApplicationStatus{},
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	app, err = s.repo.Create(ctx, app)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create application %+v", err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{Application: fromK8s(app)},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	Application
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

// Get implements the HTTP handler for getting applications
//
// @name: GetApplication
// @description: Get returns a given application
// @action: application:get
// @resource: application:{applicationID}
// @endpoint: GET /applications/{applicationID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	applicationID := mux.Vars(req)["applicationID"]
	if applicationID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing application name"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	app, err := s.repo.Get(ctx, model.AccountNamespaceName(accountID), applicationID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Application with name %s does not exist", applicationID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			Application: fromK8s(app),
		},
	}, nil

}

// ListResponse represents the response type
type ListResponse struct {
	Applications []Application `json:"applications"`
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

// List implements the HTTP handler to list applications
//
// @name: ListApplications
// @description: list all applications
// @action: application:list
// @resource: application:*
// @endpoint: GET /applications
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

	apps, err := s.repo.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get application list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	result := make([]Application, len(apps))
	for i, app := range apps {
		result[i] = fromK8s(app)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Applications: result,
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

// Delete implements the HTTP handler to delete an application
//
// @name: DeleteApplication
// @description: Deletes an application
// @action: application:delete
// @resource: application:{applicationID}
// @endpoint: DELETE /applications/{applicationID}
// @middlewares: auth
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	applicationID := mux.Vars(req)["applicationID"]
	if applicationID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing application name"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.repo.Get(ctx, model.AccountNamespaceName(accountID), applicationID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Application with name %s does not exist", applicationID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	deployList, err := s.deploymentRepo.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get deployment list to check dependencies: %+v", err),
		)
	}

	for _, d := range deployList {
		if d.Spec.App == applicationID {
			return nil, errors.E(
				errors.Internal,
				fmt.Errorf("Cannot delete application %s as there are deployments bound to it", applicationID),
			)
		}
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	err = s.repo.Delete(ctx, model.AccountNamespaceName(accountID), applicationID)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete application %s: %+v", applicationID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil
}
