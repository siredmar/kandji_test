package v20181204

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api-types"
	v20181128Deployment "github.com/grid-x/ds-api-types/management/2018-11-28/deployments"
)

const (
	// MaintenanceTaskTypeRestart represents a task to restart a device
	MaintenanceTaskTypeRestart MaintenanceTaskType = "Restart"
	// MaintenanceTaskTypeShutdown represents a task to shutdown a device
	MaintenanceTaskTypeShutdown MaintenanceTaskType = "Shutdown"
)

// Task is the maintenance task as represented by this API version
type Task struct {
	Metadata types.Metadata        `json:"metadata,omitempty"`
	Spec     MaintenanceTaskSpec   `json:"spec"`
	Status   MaintenanceTaskStatus `json:"status"`
}

// MaintenanceTaskSpec defines the desired state of MaintenanceTask
type MaintenanceTaskSpec struct {
	Type     MaintenanceTaskType          `json:"type"`
	Selector v20181128Deployment.Selector `json:"selector"`
}

// MaintenanceTaskStatus defines the observed state of MaintenanceTask
type MaintenanceTaskStatus struct {
	//+optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`
	//+optional
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`
	Successful int          `json:"successful"`
	Failed     int          `json:"failed"`
	Running    int          `json:"running"`
}

// MaintenanceTaskType indicates the type of the task
type MaintenanceTaskType string
