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

	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/types"
	v20181203 "github.com/grid-x/ds-api/types/device/2018-12-03/pod"
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
	ListByDeviceID(ctx context.Context, namespace string, deviceID string, unfiltered bool) ([]*corev1beta1.DevicePod, error)
	Get(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error)
	UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)
}

func podFromK8s(pod *corev1beta1.DevicePod) *v20181203.Pod {
	return &v20181203.Pod{
		Metadata: types.ConvertFromK8sMetadata(pod.ObjectMeta, true),
		Spec: v20181203.DevicePodSpec{
			DeviceID: pod.Spec.DeviceID,
			Config: v20181203.PodConfig{
				Volumes:                       volumesFromK8s(pod.Spec.Config.Volumes),
				Containers:                    containersFromK8s(pod.Spec.Config.Containers),
				Network:                       networkSettingsFromK8s(pod.Spec.Config.Network),
				RestartPolicy:                 restartPolicyFromK8s(pod.Spec.Config.RestartPolicy),
				TerminationGracePeriodSeconds: pod.Spec.Config.TerminationGracePeriodSeconds,
				Priority:                      pod.Spec.Config.Priority,
			},
		},
		Status: statusFromK8s(pod.Status),
	}
}

func networkSettingsFromK8s(n corev1beta1.NetworkSetting) v20181203.NetworkSetting {
	return v20181203.NetworkSetting(string(n))
}

func restartPolicyFromK8s(pol *corev1beta1.RestartPolicy) *v20181203.RestartPolicy {
	if pol == nil {
		return nil
	}

	p := v20181203.RestartPolicy{}
	p.MaxRestartCount = pol.MaxRestartCount
	p.Type = v20181203.RestartPolicyType(string(pol.Type))

	return &p
}

func volumesFromK8s(vol []corev1beta1.Volume) []v20181203.Volume {
	if len(vol) == 0 {
		return nil
	}

	volumes := make([]v20181203.Volume, len(vol))
	for i, e := range vol {
		v := v20181203.Volume{}
		v.Name = e.Name

		if e.VolumeSource.HostPath != nil {
			t := v20181203.HostPathType(string(*e.VolumeSource.HostPath.Type))
			source := v20181203.VolumeSource{
				HostPath: &v20181203.HostPathVolumeSource{
					Path: e.VolumeSource.HostPath.Path,
					Type: &t,
				},
			}

			v.VolumeSource = source
		}
		volumes[i] = v
	}
	return volumes
}

func containersFromK8s(con []corev1beta1.Container) []v20181203.Container {
	if len(con) == 0 {
		return nil
	}

	container := make([]v20181203.Container, len(con))
	for i, e := range con {
		c := v20181203.Container{
			Name:       e.Name,
			Image:      e.Image,
			Command:    e.Command,
			Args:       e.Args,
			WorkingDir: e.WorkingDir,
		}

		if e.LivelinessProbe != nil {
			probe := &v20181203.Probe{
				InitialDelaySeconds: e.LivelinessProbe.InitialDelaySeconds,
				TimeoutSeconds:      e.LivelinessProbe.TimeoutSeconds,
				PeriodSeconds:       e.LivelinessProbe.PeriodSeconds,
				SuccessThreshold:    e.LivelinessProbe.SuccessThreshold,
				FailureThreshold:    e.LivelinessProbe.FailureThreshold,
			}

			handler := v20181203.Handler{}
			if e.LivelinessProbe.Handler.Exec != nil {
				e := &v20181203.ExecAction{
					Command: e.LivelinessProbe.Handler.Exec.Command,
				}
				handler.Exec = e
			}
			if e.LivelinessProbe.Handler.HTTPGet != nil {
				g := &v20181203.HTTPGetAction{
					Path:   e.LivelinessProbe.Handler.HTTPGet.Path,
					Port:   e.LivelinessProbe.Handler.HTTPGet.Port,
					Host:   e.LivelinessProbe.Handler.HTTPGet.Host,
					Scheme: v20181203.URIScheme(string(e.LivelinessProbe.Handler.HTTPGet.Scheme)),
				}

				if len(e.LivelinessProbe.Handler.HTTPGet.HTTPHeaders) != 0 {
					httpHeaders := make([]v20181203.HTTPHeader, len(e.LivelinessProbe.Handler.HTTPGet.HTTPHeaders))
					for _, header := range e.LivelinessProbe.Handler.HTTPGet.HTTPHeaders {
						h := v20181203.HTTPHeader{
							Name:  header.Name,
							Value: header.Value,
						}

						httpHeaders[i] = h
					}
					g.HTTPHeaders = httpHeaders
				}

				handler.HTTPGet = g
			}
			probe.Handler = handler

			c.LivelinessProbe = probe
		}

		if e.ReadinessProbe != nil {
			probe := &v20181203.Probe{
				InitialDelaySeconds: e.ReadinessProbe.InitialDelaySeconds,
				TimeoutSeconds:      e.ReadinessProbe.TimeoutSeconds,
				PeriodSeconds:       e.ReadinessProbe.PeriodSeconds,
				SuccessThreshold:    e.ReadinessProbe.SuccessThreshold,
				FailureThreshold:    e.ReadinessProbe.FailureThreshold,
			}

			handler := v20181203.Handler{}
			if e.ReadinessProbe.Handler.Exec != nil {
				e := &v20181203.ExecAction{
					Command: e.ReadinessProbe.Handler.Exec.Command,
				}
				handler.Exec = e
			}
			if e.ReadinessProbe.Handler.HTTPGet != nil {
				g := &v20181203.HTTPGetAction{
					Path:   e.ReadinessProbe.Handler.HTTPGet.Path,
					Port:   e.ReadinessProbe.Handler.HTTPGet.Port,
					Host:   e.ReadinessProbe.Handler.HTTPGet.Host,
					Scheme: v20181203.URIScheme(string(e.ReadinessProbe.Handler.HTTPGet.Scheme)),
				}

				if len(e.ReadinessProbe.Handler.HTTPGet.HTTPHeaders) != 0 {
					httpHeaders := make([]v20181203.HTTPHeader, len(e.ReadinessProbe.Handler.HTTPGet.HTTPHeaders))
					for i, header := range e.ReadinessProbe.Handler.HTTPGet.HTTPHeaders {
						h := v20181203.HTTPHeader{
							Name:  header.Name,
							Value: header.Value,
						}

						httpHeaders[i] = h
					}
					g.HTTPHeaders = httpHeaders
				}

				handler.HTTPGet = g
			}
			probe.Handler = handler

			c.ReadinessProbe = probe
		}

		if len(e.Environment) != 0 {
			envVars := make([]v20181203.EnvVar, len(e.Environment))
			for i, env := range e.Environment {
				e := v20181203.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				}

				envVars[i] = e
			}
			c.Environment = envVars
		}

		if len(e.VolumeMounts) != 0 {
			volumeMounts := make([]v20181203.VolumeMount, len(e.VolumeMounts))
			for i, vol := range e.VolumeMounts {
				v := v20181203.VolumeMount{
					Name:      vol.Name,
					ReadOnly:  vol.ReadOnly,
					MountPath: vol.MountPath,
					SubPath:   vol.SubPath,
				}

				volumeMounts[i] = v
			}
			c.VolumeMounts = volumeMounts
		}

		if len(e.Ports) != 0 {
			ports := make([]v20181203.ContainerPort, len(e.Ports))
			for i, port := range e.Ports {
				p := v20181203.ContainerPort{
					Name:          port.Name,
					HostPort:      port.HostPort,
					ContainerPort: port.ContainerPort,
				}

				ports[i] = p
			}
			c.Ports = ports
		}

		container[i] = c
	}
	return container
}

func statusFromK8s(s corev1beta1.DevicePodStatus) v20181203.DevicePodStatus {
	return v20181203.DevicePodStatus{
		StartTime: s.StartTime,
	}
}

func statusToK8s(s v20181203.DevicePodStatus) corev1beta1.DevicePodStatus {
	return corev1beta1.DevicePodStatus{
		StartTime: s.StartTime,
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
type ListResponse v20181203.ListResponse

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

	ps, err := s.podClient.ListByDeviceID(ctx, model.AccountNamespaceName(accountID), deviceID, true)
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

	result := make([]*v20181203.Pod, len(ps))
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
type UpdateRequest v20181203.UpdateRequest

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

	p, err := s.podClient.Get(ctx, model.AccountNamespaceName(accountID), podID, true)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get pod: %+v", err),
		)
	}

	p.Status = statusToK8s(*payload.Status)

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
