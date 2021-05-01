package v20210310

import (
	types "github.com/grid-x/ds-api-types"
)

// DeviceConfigMap as exposed by this API version
type DeviceConfigMap struct {
	Metadata types.Metadata        `json:"metadata,omitempty"`
	Spec     DeviceConfigMapSpec   `json:"spec"`
	Status   DeviceConfigMapStatus `json:"status"`
}

// DeviceConfigMapSpec defines the desired state of DeviceConfigMap
type DeviceConfigMapSpec struct {
	// Immutable field, if set, ensures that data stored in the ConfigMap cannot
	// be updated (only object metadata can be modified).
	// +optional
	Immutable *bool `json:"immutable,omitempty"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// BinaryData contains the binary data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// BinaryData can contain byte sequences that are not in the UTF-8 range.
	// The keys stored in BinaryData must not overlap with the ones in
	// the Data field
	// +optional
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}

// DeviceConfigMapStatus defines the observed state of DeviceConfigMap
type DeviceConfigMapStatus struct{}
