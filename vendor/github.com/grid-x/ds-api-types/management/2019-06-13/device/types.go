package v20190613

import (
	types "github.com/grid-x/ds-api-types"
)

// Device as used by this API version
type Device struct {
	Metadata types.Metadata `json:"metadata"`
	Spec     DeviceSpec     `json:"spec,omitempty"`
	Status   DeviceStatus   `json:"status,omitempty"`
}

// DeviceSpec represents the spec of a device
type DeviceSpec struct {
	// The Serialnumber of the device
	Serialnumber string `json:"serialnumber"`
	// The deviceID mender assigned to this device. This can later be used
	// to automate deployment
	MenderDeviceID *string `json:"menderDeviceID,omitempty"`
	// gridX internal device ID
	InternalDeviceID *string `json:"internalDeviceID,omitempty"`
	// The MAC Address of the device assigned during provisioning
	MACAddress *string `json:"macAddress,omitempty"`
	// The public key of the device in PEM format
	PublicKey *string `json:"publicKey,omitempty"`
	// Defines a weekly time window of the form Sun:04:00-Sun:06:00
	MaintenanceWindow *types.MaintenanceWindow `json:"maintenanceWindow,omitempty"`
	// Architecture is the architecture of the device eg. arm32v7 or arm64v8
	Architecture *string `json:"architecture,omitempty"`
	// The time when the device was first seen online after being provisioned
	// This field will be set by the api after receiving the first status update outside of the provisioning process
	// In case the device has not been online yet this will be null
	FirstSeen *types.Time `json:"firstSeen,omitempty"`
	// Additional identifier to be used e.g. during commissioning on application layer
	Startcode *string `json:"startcode,omitempty"`
	// Specifies the complete networking configuration of a device.
	NetworkConfig types.NetworkConfig `json:"networkConfig,omitempty"`
}

// UpdateSpec represents the update spec type
type UpdateSpec struct {
	MACAddress        *string                  `json:"macAddress,omitempty"`
	MaintenanceWindow *types.MaintenanceWindow `json:"maintenanceWindow,omitempty"`
	MenderDeviceID    *string                  `json:"menderDeviceID,omitempty"`
	NetworkConfig     types.NetworkConfig      `json:"networkConfig,omitempty"`
}

// DeviceStatus represents the status of a device
type DeviceStatus struct {
	// The time when the device was first seen online after being provisioned
	// This field will be set by the api after receiving the first status update outside of the provisioning process
	// In case the device has not been online yet this will be null
	FirstSeen *types.Time `json:"firstSeen,omitempty"`
	// The time of the last heartbeat
	// In case the device never contacted us this will be null
	LastHeartbeat *types.Time `json:"lastHeartbeat,omitempty"`
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
	LastHeartbeatTime types.Time `json:"lastHeartbeatTime,omitempty"`
	// LastTransitionTime provides a timestamp for when the Pod last
	// transitioned from one status to another.
	LastTransitionTime *types.Time `json:"lastTransitionTime,omitempty"`
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
	// Status is a rough indication of this interface's status, if possible to determine.
	Status *NetworkInterfaceStatus `json:"status,omitempty"`
	// If true, this interface was last used to reach the DS API.
	DSUsed bool `json:"dsUsed,omitempty"`
}

// NetworkInterfaceStatus is a rough indication of this interface's status.
type NetworkInterfaceStatus string

const (
	// NetworkInterfaceStatusUnavailable means that some hard precondition for using this interface is not fulfilled, e.g. the Ethernet cable is not plugged.
	NetworkInterfaceStatusUnavailable NetworkInterfaceStatus = "Unavailable"
	// NetworkInterfaceStatusDisconnected means that the interface can possibly be used, but is not connected right now.
	NetworkInterfaceStatusDisconnected NetworkInterfaceStatus = "Disconnected"
	// NetworkInterfaceStatusActivating is reported while e.g. DHCP requests are being made, Modem is dialed...
	NetworkInterfaceStatusActivating NetworkInterfaceStatus = "Activating"
	// NetworkInterfaceStatusActivated is reported for an interface that is ready to use.
	NetworkInterfaceStatusActivated NetworkInterfaceStatus = "Activated"
	// NetworkInterfaceStatusDeactivating is repoted if the interface was deactivated by the user, or because some
	// earlier step failed.
	NetworkInterfaceStatusDeactivating NetworkInterfaceStatus = "Deactivating"
)

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
	// PublicIP is the public IP of this device
	PublicIP *string `json:"publicIP,omitempty"`
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
