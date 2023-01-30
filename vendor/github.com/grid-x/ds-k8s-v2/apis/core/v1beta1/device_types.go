/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeviceSpec defines the desired state of Device
type DeviceSpec struct {
	// The AccountID the device belongs to (the namespace is also dependent on this id)
	AccountID string `json:"accountID"`
	// the Serialnumber of the device
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`[A-Z][0-9]{3}-[0-9]{3}-[0-9]{3}-[0-9]{3}-[0-9]{3}-(A|B|P|Z)-X`
	Serialnumber string `json:"serialnumber"`
	// The deviceID mender assigned to this device. This can later be used
	// to automate deployment
	// +optional
	MenderDeviceID *string `json:"menderDeviceID,omitempty"`
	// gridX internal device ID
	// +optional
	InternalDeviceID *string `json:"internalDeviceID,omitempty"`
	// The public key of the device in PEM format
	// +optional
	PublicKey *string `json:"publicKey,omitempty"`
	// The MAC Address of the device assigned during provisioning
	// +optional
	MACAddress *string `json:"macAddress,omitempty"`
	// Defines a weekly time window of the form Sun:04:00-Sun:06:00
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
	// Architecture is the architecture of the device
	// +kubebuilder:validation:Enum=arm32v7;arm64v8;amd64
	Architecture *string `json:"architecture,omitempty"`
	// The time when the device was first seen online after being provisioned
	// This field will be set by the api after receiving the first status update outside of the provisionig process
	// In case the device has not been online yet this will be null
	// +optional
	FirstSeen *metav1.Time `json:"firstSeen,omitempty"`
}

// DeviceStatus defines the observed state of Device
type DeviceStatus struct {
	// The time when the device was first seen online after being provisioned
	// This field will be set by the api after receiving the first status update outside of the provisionig process
	// In case the device has not been online yet this will be null
	// +optional
	FirstSeen *metav1.Time `json:"firstSeen,omitempty"`
	// The time of the last heartbeat
	// In case the device never contacted us this will be null
	// +optional
	LastHeartbeat *metav1.Time `json:"lastHeartbeat,omitempty"`
	// Capacity represents the total capacity of this device
	// +optional
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocated represents the resources that are allocated
	// +optional
	Allocated ResourceList `json:"allocated,omitempty"`
	// Conditions
	// +optional
	Conditions []DeviceCondition `json:"conditions,omitempty"`
	// Info contains the system info of the device
	// +optional
	Info *DeviceSystemInfo `json:"info,omitempty"`
	// Images
	// +optional
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
	// condition
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// LastTransitionTime provides a timestamp for when the Pod last
	// transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason is a unique, one-word, CamelCase reason for the condition’s
	// last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about the
	// transition.
	// +optional
	Message *string `json:"message,omitempty"`
}

// NetworkInterface represents information about a network interface
type NetworkInterface struct {
	// Name is the name of the interface, e.g. eth0
	Name string `json:"name"`
	// MACAddress is the mac address of this interface
	MACAddress string `json:"macAddress,omitempty"`
	// IPv4 is the IP v4 address of this interface
	// +optional
	IPv4 *string `json:"ipV4,omitempty"`
	// IPVv6 is the IP v6 address of this inteface
	// +optional
	IPv6 *string `json:"ipV6,omitempty"`
}

// DeviceSystemInfo represents the system info of a device
type DeviceSystemInfo struct {
	// Hostname is the hostname of the device
	// +optional
	Hostname *string `json:"hostname,omitempty"`
	// Architecture is the architecture of the device
	// +optional
	Architecture *string `json:"architecture,omitempty"`
	// HardwareVersion is the hardware version (e.g. v2)
	// +optional
	HardwareVersion *string `json:"hardwareVersion,omitempty"`
	// CPUModel is the cpu model of the device (see /proc/cpuinfo)
	// +optional
	CPUModel *string `json:"cpuModel,omitempty"`
	// OSVersion is the version of the OS
	// +optional
	OSVersion *string `json:"osVersion,omitempty"`
	// KernelVersion is the version of the kernel
	// +optional
	KernelVersion *string `json:"kernelVersion,omitempty"`
	// SupervisorVersion is the version of the supervisor
	// +optional
	SupervisorVersion *string `json:"supervisorVersion,omitempty"`
	// DockerVersion is the version of docker
	// +optional
	DockerVersion *string `json:"dockerVersion,omitempty"`
	// LastReboot is the time the last reboot happened (run `last reboot`
	// for this)
	// +optional
	LastReboot *string `json:"lastReboot,omitempty"`
	// BootID is the current boot ID (see /proc/sys/kernel/random/boot_id)
	BootID *string `json:"bootID,omitempty"`
	// MachineID is the machine ID of this device (see /etc/machine-id)
	MachineID *string `json:"machineID,omitempty"`
	// NetworkInterfaces is the list of network interfaces of this device
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces,omitempty"`
	// PublicIP contains the public IP of the device
	PublicIP *string `json:"publicIP,omitempty"`
}

// ContainerImage represents a container image with different names and its
// size in bytes
type ContainerImage struct {
	// Names lists the different names this image has
	Names     []string `json:"names,omitempty"`
	SizeBytes uint64   `json:"sizeBytes,omitempty"`
}

// +kubebuilder:object:root=true
// Device is the Schema for the devices API
// +kubebuilder:subresource:status
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DeviceList contains a list of Device
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Device{}, &DeviceList{})
}
