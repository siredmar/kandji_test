package v20181127

// CreateRequest represents the request type
type CreateRequest struct {
	Name string `json:"name"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	Application
}

// GetResponse represents the response type
type GetResponse struct {
	Application
}

// ListResponse represents the response type
type ListResponse struct {
	Applications []Application `json:"applications"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
