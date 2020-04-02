package v20181203

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "github.com/grid-x/ds-api-types"
)

// DevicePodSpec defines the desired state of DevicePod
type DevicePodSpec struct {
	// DeviceID is the ID of the device this pod should run on
	DeviceID string `json:"deviceID"`
	// Config provides the configuration of the pod, i.e. all containers
	// etc.
	Config PodConfig `json:"config"`
}

// PodConfig describes the config of the pod
type PodConfig struct {
	// List of volumes
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`
	// List of containers belonging to this pod
	Containers []Container `json:"containers"`
	// The network setting
	// +optional
	Network NetworkSetting `json:"network,omitempty"`
	// Restart policy for the pod
	// +optional
	RestartPolicy *RestartPolicy `json:"restartPolicy,omitempty"`
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// The priority value. Various system components use this field to find the
	// priority of the pod.
	// The higher the value, the higher the priority.
	// NOTE: This is currently not supported
	// +optional
	Priority *int32 `json:"priority,omitempty"`
}

// DevicePodStatus defines the observed state of DevicePod
type DevicePodStatus struct {
	// Date and time at which the object was acknowledged by the supervisor.
	// This is before the supervisor pulled the container image(s) for the
	// pod.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Conditions of the pod
	// +optional
	Conditions []PodCondition `json:"conditions,omitempty"`
	// Detailed status of each container in the pod
	// +optional
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
}

// ContainerStateWaiting represents the state when a container is waiting
type ContainerStateWaiting struct {
	// A brief CamelCase string indicating details about why the container
	// is in waiting state.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human-readable message indicating details about why the container
	// is in waiting state.
	// +optional
	Message string `json:"message,omitempty"`
}

// ContainerStateRunning represents the state when a container is running
type ContainerStateRunning struct {
	// +optional
	StartedAt metav1.Time `json:"startedAt,omitempty"`
}

// ContainerStateTerminated represents the state when a container has terminated
type ContainerStateTerminated struct {
	ExitCode int32 `json:"exitCode,omitempty"`
	// +optional
	Signal int32 `json:"signal,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	StartedAt metav1.Time `json:"startedAt,omitempty"`
	// +optional
	FinishedAt metav1.Time `json:"finishedAt,omitempty"`
	// +optional
	ContainerID string `json:"containerID,omitempty"`
}

// ContainerState holds a possible state of container.
// Only one of its members may be specified.
// If none of them is specified, the default one is ContainerStateWaiting.
type ContainerState struct {
	// +optional
	Waiting *ContainerStateWaiting `json:"waiting,omitempty"`
	// +optional
	Running *ContainerStateRunning `json:"running,omitempty"`
	// +optional
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

// ContainerStatus represents the status of an individual container
type ContainerStatus struct {
	// Each container in a pod must have a unique name.
	Name string `json:"name"`
	// +optional
	State ContainerState `json:"state,omitempty"`
	// +optional
	LastTerminationState ContainerState `json:"lastTerminationState,omitempty"`
	// Ready specifies whether the container has passed its readiness check.
	Ready bool `json:"ready"`
	// Note that this is calculated from dead containers.  But those containers are subject to
	// garbage collection.  This value will get capped at 5 by GC.
	RestartCount int32  `json:"restartCount"`
	Image        string `json:"image"`
	ImageID      string `json:"imageID"`
	// +optional
	ContainerID string `json:"containerID,omitempty"`
}

// PodConditionType is the type of a pod condition
type PodConditionType string

// These are valid conditions of pod.
const (
	// PodConditionScheduled means the pod was received by the supervisor
	// and is scheduled for execution. This does not mean that the pod is
	// ready to start and can also mean that we need to download the
	// images first. Check other conditions for details.
	PodConditionScheduled PodConditionType = "Scheduled"
	// PodConditionDownloading indicates that the supervisor currently
	// downloads the images for all the containers.
	PodConditionDownloading PodConditionType = "Downloading"
	// PodConditionReady indicates that the pod is running.
	PodConditionReady PodConditionType = "Running"
	// PodConditionCompleted indicates that the pod completed, i.e. if it
	// is set to run to completion and all containers exited it will have
	// this condition.
	PodConditionCompleted PodConditionType = "Completed"
	// PodConditionFailed indicates that at least one container exited with
	// a non-zero exit status. Depending on the RestartPolicy the pod will
	// be restarted.
	PodConditionFailed PodConditionType = "Failed"
)

// PodCondition represents a condition of a pod
type PodCondition struct {
	Type   PodConditionType `json:"type"`
	Status ConditionStatus  `json:"status"`
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

// NetworkSetting is an enum value that describes how the network setting for
// a pod should be set
type NetworkSetting string

// These are the currently supported network settings
const (
	// NetworkSettingNone is the default value
	NetworkSettingNone NetworkSetting = "None"
	// NetworkSettingHost uses the host network for the whole pod
	NetworkSettingHost NetworkSetting = "Host"
	// NetworkSettingBridge uses the bridge network setup for the pod
	NetworkSettingBridge NetworkSetting = "Bridge"
)

// RestartPolicy describes the restart policy of a pod
type RestartPolicy struct {
	// RestartPolicyType indicates which type this restart policy uses
	Type RestartPolicyType `json:"type"`
	// MaxRestartCount limits the number of restarts if policy is Always
	// if set to a value less or equal to 0 it will restart indefinitely
	// +optional
	MaxRestartCount *int32 `json:"maxRestartCount,omitempty"`
}

// RestartPolicyType is the type of the restart policy
type RestartPolicyType string

// These are valid values for a restart
const (
	// RestartPolicyAlways means the pod should always be restarted
	RestartPolicyAlways RestartPolicyType = "Always"
	// RestartPolicyOnFailure means the pod will only be restarted on
	// failure
	RestartPolicyOnFailure RestartPolicyType = "OnFailure"
	// RestartPolicyNever means the pod will never be restarted
	RestartPolicyNever RestartPolicyType = "Never"
)

// EnvVar describes an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Volume describes a volume
type Volume struct {
	Name string `json:"name"`
	VolumeSource
}

// HostPathType describes the type of a host path in a volume
type HostPathType string

const (
	// HostPathDirectoryOrCreate means that if nothing exists at the given
	// path, an empty directory will be created there as needed with file
	// mode 0755, having the same group and ownership with the supervisor.
	HostPathDirectoryOrCreate HostPathType = "DirectoryOrCreate"
	// HostPathDirectory means a directory must exist at the given path
	HostPathDirectory HostPathType = "Directory"
	// HostPathFileOrCreate means that if nothing exists at the given path,
	// an empty file will be created there as needed with file mode 0644,
	// having the same group and ownership with the supervisor.
	HostPathFileOrCreate HostPathType = "FileOrCreate"
	// HostPathFile means a file must exist at the given path
	HostPathFile HostPathType = "File"
	// HostPathSocket means a UNIX socket must exist at the given path
	HostPathSocket HostPathType = "Socket"
	// HostPathCharDev means a character device must exist at the given path
	HostPathCharDev HostPathType = "CharDevice"
	// HostPathBlockDev means a block device must exist at the given path
	HostPathBlockDev HostPathType = "BlockDevice"
)

// VolumeSource represents the source location of a volume to mount.
// Only one of its members may be specified.
type VolumeSource struct {
	// HostPath represents file or directory on the host machine that is
	// directly exposed to the container.
	// +optional
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty"`
}

// HostPathVolumeSource represents a host path mapped into a pod.
type HostPathVolumeSource struct {
	// If the path is a symlink, it will follow the link to the real path.
	Path string `json:"path"`
	// Defaults to ""
	// +optional
	Type *HostPathType `json:"type,omitempty"`
}

// HTTPHeader describes a custom header to be used in HTTP probes
type HTTPHeader struct {
	// The header field name
	Name string `json:"name"`
	// The header field value
	Value string `json:"value"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// Optional: Path to access on the HTTP server.
	// +optional
	Path string `json:"path,omitempty"`
	// Required: Number of the port to access on the container.
	// +optional
	Port *int32 `json:"port,omitempty"`
	// Optional: Host name to connect to, defaults to the pod IP. You
	// probably want to set "Host" in httpHeaders instead.
	// +optional
	Host string `json:"host,omitempty"`
	// Optional: Scheme to use for connecting to the host, defaults to HTTP.
	// +optional
	Scheme URIScheme `json:"scheme,omitempty"`
	// Optional: Custom headers to set in the request. HTTP allows repeated headers.
	// +optional
	HTTPHeaders []HTTPHeader `json:"headers,omitempty"`
}

// URIScheme identifies the scheme used for connection to a host for Get actions
type URIScheme string

const (
	// URISchemeHTTP means that the scheme used will be http://
	URISchemeHTTP URIScheme = "HTTP"
	// URISchemeHTTPS means that the scheme used will be https://
	URISchemeHTTPS URIScheme = "HTTPS"
)

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem.  The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work.  To use
	// a shell, you need to explicitly call out to that shell.
	// +optional
	Command []string `json:"command,omitempty"`
}

// HealthCheck describes docker healthchecks
type HealthCheck struct {
	// HealthMonitoringEnabled describes whether or not a failing healthcheck
	// should be handled.
	// +optional
	HealthMonitoringEnabled bool `json:"healthMonitoringEnabled,omitempty"`
	// +optional
	Config *HealthConfig `json:"config,omitempty"`
}

// HealthConfig describes the config of a docker healthcheck
type HealthConfig struct {
	// Exec is the test to perform to check that the container is healthy.
	// +optional
	Exec *ExecAction `json:"exec,omitempty"`
	// +optional
	IntervalSeconds int32 `json:"intervalSeconds,omitempty"` // Interval is the time to wait between checks.
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"` // TimeoutSeconds is the time to wait before considering the check to have hung.
	// +optional
	StartPeriodSeconds int32 `json:"periodSeconds,omitempty"` // The start period for the container to initialize before the retries starts to count down.
	// Retries is the number of consecutive failures needed to consider a container as unhealthy.
	// Zero means inherit.
	// +optional
	Retries int `json:"retries,omitempty"`
}

// ContainerPort represents a network port in a single container
type ContainerPort struct {
	// Optional: If specified, this must be an IANA_SVC_NAME  Each named port
	// in a pod must have a unique name.
	// +optional
	Name string `json:"name,omitempty"`
	// Optional: If specified, this must be a valid port number, 0 < x < 65536.
	// If HostNetwork is specified, this must match ContainerPort.
	// +optional
	HostPort int32 `json:"hostPort,omitempty"`
	// Required: This must be a valid port number, 0 < x < 65536.
	ContainerPort int32 `json:"containerPort"`
}

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	// Required: This must match the Name of a Volume [above].
	Name string `json:"name"`
	// Optional: Defaults to false (read-write).
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
	// Required. If the path is not an absolute path (e.g. some/path) it
	// will be prepended with the appropriate root prefix for the operating
	// system.  On Linux this is '/', on Windows this is 'C:\'.
	MountPath string `json:"mountPath"`
	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

// Container is the specification of a container
type Container struct {
	// The name of the container
	Name string `json:"name"`
	// The image to use for the container
	Image string `json:"image"`
	// The docker images entrypoint is used if this is not provided
	// +optional
	Command []string `json:"command,omitempty"`
	// The docker images cmd is used if this is not provided
	// +optional
	Args []string `json:"args,omitempty"`
	// The list of environment variables
	// +optional
	Environment []EnvVar `json:"env,omitempty"`
	// +optional
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
	// Workingdir overwrites the working dir set in the docker image
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`
	// +optional
	Ports []ContainerPort `json:"ports,omitempty"`
	// HealthCheck is the probe to be used to check if a container is still living
	// +optional
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

// Pod definition as exported by this API version
type Pod struct {
	Metadata types.Metadata  `json:"metadata"`
	Spec     DevicePodSpec   `json:"spec"`
	Status   DevicePodStatus `json:"status"`
}

// ConditionStatus indicates the status of a condition
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition;
// "ConditionFalse" means a resource is not in the condition; "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)
