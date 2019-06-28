package v20190529

// CreateRequest represents the request type
type CreateRequest struct {
	Spec *CleanupConfigSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*CleanupConfig
}

// GetResponse represents the response type
type GetResponse struct {
	*CleanupConfig
}

// ListResponse represents the response type
type ListResponse struct {
	CleanupConfig []*CleanupConfig `json:"cleanupConfigs"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
