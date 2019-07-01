package v20190609

// GetResponse represents the response type
type GetResponse struct {
	*DeviceDockerConfig
}

// ListResponse represents the response type
type ListResponse struct {
	DeviceDockerConfigs []*DeviceDockerConfig `json:"deviceDockerConfigs"`
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *DeviceDockerConfigStatus `json:"status,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*DeviceDockerConfig
}
