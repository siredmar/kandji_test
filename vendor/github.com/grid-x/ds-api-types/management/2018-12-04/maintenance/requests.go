package v20181204

// CreateRequest represents the request type
type CreateRequest struct {
	Spec *MaintenanceTaskSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Task
}

// ListResponse represents the response type
type ListResponse struct {
	Tasks []*Task `json:"maintenanceTasks"`
}

// GetResponse represents the response type
type GetResponse struct {
	*Task
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
