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
	version = "2018-12-04"
	group   = "pods"
)

var (
	// defaultTimeout is the default timeout used when calling downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
}

type podClient interface {
	List(ctx context.Context, namespace string) ([]*corev1beta1.DevicePod, error)
	Get(ctx context.Context, namespace, name string) (*corev1beta1.DevicePod, error)
}

// Pod as exposed by this API version
type Pod struct {
	Metadata api.Metadata                `json:"metadata,omitempty"`
	Spec     corev1beta1.DevicePodSpec   `json:"spec"`
	Status   corev1beta1.DevicePodStatus `json:"status"`
}

func podFromK8s(pod *corev1beta1.DevicePod) *Pod {
	return &Pod{
		Metadata: api.ConvertFromK8sMetadata(pod.ObjectMeta),
		Spec:     pod.Spec,
		Status:   pod.Status,
	}
}

// Service implements the read only pod HTTP handlers for the management API
type Service struct {
	logger    log.FieldLogger
	podClient podClient
	auth      authProvider
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

// List implements the HTTP handler to list pods
//
// @name: ListPods
// @description: Lists the device pods in this account
// @action: pods:List
// @resource: pods:*
// @endpoint: GET /pods
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

	ps, err := s.podClient.List(ctx, model.AccountNamespaceName(accountID))
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

// GetResponse represents the response type
type GetResponse struct {
	*Pod
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

// Get implements the HTTP handler to get pods
//
// @name: GetPod
// @description: Gets a specified pod
// @action: pods:Get
// @resource: pods:{podID}
// @endpoint: GET /pods/{podID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
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
			errors.NotExists,
			fmt.Errorf("Pod with ID %s does not exist", podID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			Pod: podFromK8s(p),
		},
	}, nil
}
