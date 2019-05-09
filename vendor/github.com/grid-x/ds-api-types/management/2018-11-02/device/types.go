package v20181102

import (
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

// DeviceStatus represents the status of a device
type DeviceStatus struct {
	LastHeartbeat string `json:"lastHeartbeat,omitempty"`
}

// UpdateSpec represents the update spec type
type UpdateSpec struct {
	MACAddress        *string `json:"macAddress,omitempty"`
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`
}
