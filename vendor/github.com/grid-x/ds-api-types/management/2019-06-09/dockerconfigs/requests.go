package v20190609

// CreateRequest represents the request type
type CreateRequest struct {
	Spec *DockerConfigSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*DockerConfig
}

// GetResponse represents the response type
type GetResponse struct {
	*DockerConfig
}

// ListResponse represents the response type
type ListResponse struct {
	DockerConfigs []*DockerConfig `json:"dockerConfigs"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
