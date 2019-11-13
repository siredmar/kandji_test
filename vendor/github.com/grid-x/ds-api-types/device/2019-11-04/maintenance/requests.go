package v20191104

// ListResponse represents the response type
type ListResponse struct {
	Tasks []*MaintenanceTask `json:"maintenanceTasks"`
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Status *MaintenanceTaskStatus `json:"status,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*MaintenanceTask
}
