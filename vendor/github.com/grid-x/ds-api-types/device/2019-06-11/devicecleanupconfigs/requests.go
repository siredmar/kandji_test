package v20190611

// GetResponse represents the response type
type GetResponse struct {
	*DeviceCleanupConfig
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *DeviceCleanupConfigStatus `json:"status,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*DeviceCleanupConfig
}
