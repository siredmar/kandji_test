package pods

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-12-03"
	group   = "pods"
)

var (
	// defaultTimeout is the default timeout with which downstream services
	// are called
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	DeviceIDFromContext(context.Context) (string, error)
	AccountIDFromContext(context.Context) (string, error)
}

type podClient interface {
	ListByDeviceID(ctx context.Context, namespace string, deviceID string) ([]*corev1beta1.DevicePod, error)
	Get(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error)
	UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)
}

// Pod definition as exported by this API version
type Pod struct {
	Metadata api.Metadata                `json:"metadata"`
	Spec     corev1beta1.DevicePodSpec   `json:"spec"`
	Status   corev1beta1.DevicePodStatus `json:"status"`
}

func podFromK8s(pod *corev1beta1.DevicePod) *Pod {
	return &Pod{
		Metadata: podMetaFromK8s(pod),
		Spec:     pod.Spec,
		Status:   pod.Status,
	}
}

func podMetaFromK8s(pod *corev1beta1.DevicePod) api.Metadata {
	return api.Metadata{
		ID:          pod.Name,
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
	}
}

// Service implements the handlers for the pod device-api endpoint
type Service struct {
	logger    log.FieldLogger
	podClient podClient
	ap        authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case podClient:
			s.podClient = i
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

// ListResponse represents the response type
type ListResponse struct {
	Pods []*Pod `json:"pods"`
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

// List implements the HTTP handler for listing pods
//
// @name: ListPods
// @description: Lists the pods for a given device
// @action: pods:List
// @resource: pods:*
// @endpoint: GET /pods
// @middlewares: auth
func (s *Service) List(req *http.Request) (*encoding.Response, error) {
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

	ps, err := s.podClient.ListByDeviceID(ctx, model.AccountNamespaceName(accountID), deviceID)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get pod list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name < ps[j].Name
	})

	result := make([]*Pod, len(ps))
	for i, p := range ps {
		result[i] = podFromK8s(p)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Pods: result,
		},
	}, nil
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *corev1beta1.DevicePodStatus `json:"status,omitempty"`
}

// Validate validates an UpdateRequest
func (req *UpdateRequest) Validate() error {
	if req.Status == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
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
	*Pod
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

// Update implements the handler for updating pod statuses
//
// @name: UpdatePod
// @description: Update pod allows to update the status of a given pod
// @action: pods:Update
// @resource: pods:{podID}
// @endpoint: PATCH /pods/{podID}
// @middlewares: auth
func (s *Service) Update(req *http.Request, payload UpdateRequest) (*encoding.Response, error) {
	accountID, err := s.ap.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	podID := mux.Vars(req)["podID"]
	if podID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing pod ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	p, err := s.podClient.Get(ctx, model.AccountNamespaceName(accountID), podID)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get pod: %+v", err),
		)
	}

	p.Status = *payload.Status

	p, err = s.podClient.UpdateStatus(ctx, p)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch pod %+v", err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateResponse{
			Pod: podFromK8s(p),
		},
	}, nil
}
