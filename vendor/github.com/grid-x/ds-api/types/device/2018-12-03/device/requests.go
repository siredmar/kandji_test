package v20181203

// GetResponse represents the response type
type GetResponse struct {
	*Device
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *DeviceStatus `json:"status"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Device
}
