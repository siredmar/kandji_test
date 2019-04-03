package v20181203

// ListResponse represents the response type
type ListResponse struct {
	Pods []*Pod `json:"pods"`
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *DevicePodStatus `json:"status,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Pod
}
