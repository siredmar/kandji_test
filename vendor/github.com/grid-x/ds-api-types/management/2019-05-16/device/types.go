package v20190516

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api-types"
)

// Device as used by this API version
type Device struct {
	Metadata types.Metadata `json:"metadata"`
	Spec     DeviceSpec     `json:"spec,omitempty"`
	Status   DeviceStatus   `json:"status,omitempty"`
}

// DeviceSpec represents the spec of a device
type DeviceSpec struct {
	Serialnumber      string  `json:"serialnumber"`
	MACAddress        *string `json:"macAddress,omitempty"`
	PublicKey         *string `json:"publicKey,omitempty"`
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
}

// UpdateSpec represents the update spec type
type UpdateSpec struct {
	MACAddress        *string `json:"macAddress,omitempty"`
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
}

// DeviceStatus represents the status of a device
type DeviceStatus struct {
	// The time of the last heartbeat
	// In case the device never contacted us this will be null
	LastHeartbeat *metav1.Time `json:"lastHeartbeat,omitempty"`
	// Capacity represents the total capacity of this device
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocated represents the resources that are allocated
	Allocated ResourceList `json:"allocated,omitempty"`
	// Conditions
	Conditions []DeviceCondition `json:"conditions,omitempty"`
	// Info contains the system info of the device
	Info *DeviceSystemInfo `json:"info,omitempty"`
	// Images
	Images []ContainerImage `json:"images,omitempty"`
}

// ResourceType represents a resource type
type ResourceType string

// A resource type can be either cpu, memory or storage
const (
	ResourceTypeCPU     ResourceType = "CPU"
	ResourceTypeMemory  ResourceType = "Memory"
	ResourceTypeStorage ResourceType = "Storage"
)

// ResourceQuantity is the quantity of the resource
type ResourceQuantity int

// ResourceList maps types to their quantities
type ResourceList map[ResourceType]ResourceQuantity

// DeviceConditionType is the type of a device condition
type DeviceConditionType string

const (
	// DeviceReady indicates that the device is ready
	DeviceReady DeviceConditionType = "Ready"
	// DeviceOutOfDisk indicates that the disk is full
	DeviceOutOfDisk DeviceConditionType = "OutOfDisk"
	// DeviceMemoryPressure indicates that the device has memory pressure
	DeviceMemoryPressure DeviceConditionType = "MemoryPressure"
	// DeviceDiskPressure indicates that the device has disk pressure
	DeviceDiskPressure DeviceConditionType = "DiskPressure"
)

// DeviceCondition represents a device condition
type DeviceCondition struct {
	// Type is the type of the device condition
	Type DeviceConditionType `json:"type,omitempty"`
	// Status is the status of the condition
	Status ConditionStatus `json:"status,omitempty"`
	// LastHeartbeatTime is the timestamp for when the Pod condition was
	// last probed.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// LastTransitionTime provides a timestamp for when the Pod last
	// transitioned from one status to another.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason is a unique, one-word, CamelCase reason for the condition’s
	// last transition.
	Reason *string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about the
	// transition.
	Message *string `json:"message,omitempty"`
}

// NetworkInterface represents information about a network interface
type NetworkInterface struct {
	// Name is the name of the interface, e.g. eth0
	Name string `json:"name"`
	// MACAddress is the mac address of this interface
	MACAddress string `json:"macAddress,omitempty"`
	// IPv4 is the IP v4 address of this interface
	IPv4 *string `json:"ipV4,omitempty"`
	// IPVv6 is the IP v6 address of this inteface
	IPv6 *string `json:"ipV6,omitempty"`
}

// DeviceSystemInfo represents the system info of a device
type DeviceSystemInfo struct {
	// Hostname is the hostname of the device
	Hostname *string `json:"hostname,omitempty"`
	// Architecture is the architecture of the device
	Architecture *string `json:"architecture,omitempty"`
	// HardwareVersion is the hardware version (e.g. v2)
	HardwareVersion *string `json:"hardwareVersion,omitempty"`
	// CPUModel is the cpu model of the device (see /proc/cpuinfo)
	CPUModel *string `json:"cpuModel,omitempty"`
	// OSVersion is the version of the OS
	OSVersion *string `json:"osVersion,omitempty"`
	// KernelVersion is the version of the kernel
	KernelVersion *string `json:"kernelVersion,omitempty"`
	// SupervisorVersion is the version of the supervisor
	SupervisorVersion *string `json:"supervisorVersion,omitempty"`
	// DockerVersion is the version of docker
	DockerVersion *string `json:"dockerVersion,omitempty"`
	// LastReboot is the time the last reboot happened (run `last reboot`
	// for this)
	LastReboot *string `json:"lastReboot,omitempty"`
	// BootID is the current boot ID (see /proc/sys/kernel/random/boot_id)
	BootID *string `json:"bootID,omitempty"`
	// MachineID is the machine ID of this device (see /etc/machine-id)
	MachineID *string `json:"machineID,omitempty"`
	// NetworkInterfaces is the list of network interfaces of this device
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces,omitempty"`
}

// ContainerImage represents a container image with different names and its
// size in bytes
type ContainerImage struct {
	// Names lists the different names this image has
	Names     []string `json:"names,omitempty"`
	SizeBytes uint64   `json:"sizeBytes,omitempty"`
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
