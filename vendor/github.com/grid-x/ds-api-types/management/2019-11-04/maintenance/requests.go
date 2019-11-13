package v20191104

// CreateRequest represents the request type
type CreateRequest struct {
	Spec *MaintenanceTaskSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*MaintenanceTask
}

// ListResponse represents the response type
type ListResponse struct {
	Tasks []*MaintenanceTask `json:"maintenanceTasks"`
}

// GetResponse represents the response type
type GetResponse struct {
	*MaintenanceTask
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
