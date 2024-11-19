package v20241104

import types "github.com/grid-x/ds-api-types"

// DeviceLogsSpec defines the desired state of DeviceLogs
type DeviceLogsSpec struct {
	// Owner
	Owner string `json:"owner"`

	// LogLevel set on Device
	LogLevel string `json:"logLevel"`

	// ExpiresAt defines the lifetime of the resource
	ExpiresAt types.Time `json:"expiresAt"`
}

// CreateSpec represents the body for the create request.
type CreateSpec struct {
	LogLevel  string     `json:"logLevel,omitempty"`
	ExpiresAt types.Time `json:"expiresAt,omitempty"`
}

// UpdateSpec represents the body for the update request.
type UpdateSpec struct {
	LogLevel  string     `json:"logLevel,omitempty"`
	ExpiresAt types.Time `json:"expiresAt,omitempty"`
}

// DeviceLogsStatus defines the observed state of DeviceLogs
type DeviceLogsStatus struct {
	// NotifiedAt records successful notifications of pending resource expiration
	NotifiedAt []types.Time `json:"notifiedAt"`
}

// DeviceLogs is the Schema for the devicelogs API
type DeviceLogs struct {
	Metadata types.Metadata   `json:"metadata,omitempty"`
	Spec     DeviceLogsSpec   `json:"spec,omitempty"`
	Status   DeviceLogsStatus `json:"status,omitempty"`
}

// DeviceLogsList contains a list of DeviceLogs
type DeviceLogsList struct {
	Items []DeviceLogs `json:"items"`
}
