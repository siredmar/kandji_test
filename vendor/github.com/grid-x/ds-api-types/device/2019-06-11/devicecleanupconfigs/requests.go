package v20190611

// GetResponse represents the response type
type GetResponse struct {
	*DeviceCleanupConfig
}

// ListResponse represents the response type
type ListResponse struct {
	DeviceCleanupConfig []*DeviceCleanupConfig `json:"deviceCleanupConfigs"`
}
