package v20191104

import (
	types "github.com/grid-x/ds-api-types"
)

const (
	// MaintenanceTaskTypeRestart represents a task to restart a device
	MaintenanceTaskTypeRestart MaintenanceTaskType = "Restart"
)

// MaintenanceTask is the maintenance task as represented by this API version
type MaintenanceTask struct {
	Metadata types.Metadata        `json:"metadata,omitempty"`
	Spec     MaintenanceTaskSpec   `json:"spec"`
	Status   MaintenanceTaskStatus `json:"status"`
}

// MaintenanceTaskSpec defines the desired state of MaintenanceTask
type MaintenanceTaskSpec struct {
	Type     MaintenanceTaskType `json:"type"`
	DeviceID string              `json:"deviceID"`
}

// MaintenanceTaskStatus defines the observed state of MaintenanceTask
type MaintenanceTaskStatus struct {
	//+optional
	StartedAt *types.Time `json:"startedAt,omitempty"`
	//+optional
	FinishedAt *types.Time `json:"finishedAt,omitempty"`
	Successful int         `json:"successful"`
	Failed     int         `json:"failed"`
	Running    int         `json:"running"`
}

// MaintenanceTaskType indicates the type of the task
type MaintenanceTaskType string
