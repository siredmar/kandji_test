package v20190529

// GetResponse represents the response type
type GetResponse struct {
	*DeviceCleanupConfig
}

// ListResponse represents the response type
type ListResponse struct {
	DeviceCleanupConfig []*DeviceCleanupConfig `json:"deviceCleanupConfigs"`
}
