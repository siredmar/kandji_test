package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-12-04"
	group   = "maintenance"
)

var (
	// defaultTimeout is the default timeout that is used when calling
	// downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
}

type maintenanceTaskClient interface {
	Create(ctx context.Context, task *mainv1beta1.MaintenanceTask) (*mainv1beta1.MaintenanceTask, error)
	Delete(ctx context.Context, namespace, name string) error
	List(ctx context.Context, namespace string) ([]*mainv1beta1.MaintenanceTask, error)
	Get(ctx context.Context, namespace, name string) (*mainv1beta1.MaintenanceTask, error)
}

// Task is the maintenance task as represented by this API version
type Task struct {
	Metadata api.Metadata                      `json:"metadata,omitempty"`
	Spec     mainv1beta1.MaintenanceTaskSpec   `json:"spec"`
	Status   mainv1beta1.MaintenanceTaskStatus `json:"status"`
}

func taskFromK8s(task *mainv1beta1.MaintenanceTask) *Task {
	return &Task{
		Metadata: api.ConvertFromK8sMetadata(task.ObjectMeta),
		Spec:     task.Spec,
		Status:   task.Status,
	}
}

// Service implements the HTTP handler for creating, reading, updating and
// deleting maintenance tasks
type Service struct {
	logger   log.FieldLogger
	mtClient maintenanceTaskClient
	auth     authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case maintenanceTaskClient:
			s.mtClient = i
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
	Spec *mainv1beta1.MaintenanceTaskSpec `json:"spec"`
}

// Validate validates a CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Spec == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing spec"),
		)
	}
	if req.Spec.Type == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing type"),
		)
	}
	if req.Spec.Selector.MatchByLabels == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing labels to match"),
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
	*Task
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

// Create implements the HTTP handler to create maintenance tasks
//
// @name: CreateMaintenanceTask
// @description: Creates a new maintenance task
// @action: maintenance:Create
// @resource: maintenance:*
// @endpoint: POST /maintenance
func (s *Service) Create(req *http.Request, payload CreateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	taskID := uuid.NewV4().String()
	task := &mainv1beta1.MaintenanceTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:      taskID,
			Namespace: model.AccountNamespaceName(accountID),
		},
		Spec:   *payload.Spec,
		Status: mainv1beta1.MaintenanceTaskStatus{},
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	task, err = s.mtClient.Create(ctx, task)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create maintenance Task %s: %+v", taskID, err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{Task: taskFromK8s(task)},
	}, nil
}

// ListResponse represents the response type
type ListResponse struct {
	Tasks []*Task `json:"maintenanceTasks"`
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

// List implements the HTTP handler for listing maintenance tasks
//
// @name: ListMaintenanceTasks
// @description: Lists the existing maintenance tasks
// @action: maintenance:List
// @resource: maintenance:*
// @endpoint: GET /maintenance
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

	ms, err := s.mtClient.List(ctx, model.AccountNamespaceName(accountID))
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get maintenance task list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(ms, func(i, j int) bool {
		return ms[i].Name < ms[j].Name
	})

	result := make([]*Task, len(ms))
	for i, m := range ms {
		result[i] = taskFromK8s(m)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Tasks: result,
		},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	*Task
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

// Get implements the HTTP handler for getting maintenance tasks
//
// @name: GetMaintenanceTask
// @description: Fetches a single maintenance task
// @action: maintenance:Get
// @resource: maintenance:{taskID}
// @endpoint: GET /maintenance/{taskID}
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	taskID := mux.Vars(req)["taskID"]
	if taskID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing maintenance task id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	task, err := s.mtClient.Get(ctx, model.AccountNamespaceName(accountID), taskID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Maintenance task with name %s does not exist", taskID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			Task: taskFromK8s(task),
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

// Delete implements the HTTP handler for deleting maintenance tasks
//
// @name: DeleteMaintenanceTask
// @description: Deletes an existing maintenance task
// @action: maintenance:Delete
// @resource: maintenance:{taskID}
// @endpoint: DELETE /maintenance/{taskID}
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	taskID := mux.Vars(req)["taskID"]
	if taskID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing maintenance task id"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.mtClient.Get(ctx, model.AccountNamespaceName(accountID), taskID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Maintenance task with ID %s does not exist", taskID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	if err := s.mtClient.Delete(ctx, model.AccountNamespaceName(accountID), taskID); err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete maintenance task %s: %+v", taskID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil
}
