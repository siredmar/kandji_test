package v20190514

// GetResponse represents the response type
type GetResponse struct {
	*DeviceDockerConfig
}

// ListResponse represents the response type
type ListResponse struct {
	DeviceDockerConfigs []*DeviceDockerConfig `json:"deviceDockerConfigs"`
}
