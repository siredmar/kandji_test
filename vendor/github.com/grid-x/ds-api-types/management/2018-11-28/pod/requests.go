package v20181128

// ListResponse represents the response type
type ListResponse struct {
	Pods []*Pod `json:"pods"`
}

// GetResponse represents the response type
type GetResponse struct {
	*Pod
}
