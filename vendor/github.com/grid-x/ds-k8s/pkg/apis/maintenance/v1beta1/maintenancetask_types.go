/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package v1beta1

import (
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MaintenanceTaskType indicates the type of the task
type MaintenanceTaskType string

const (
	// MaintenanceTaskTypeRestart represents a task to restart a device
	MaintenanceTaskTypeRestart MaintenanceTaskType = "Restart"
	// MaintenanceTaskTypeShutdown represents a task to shutdown a device
	MaintenanceTaskTypeShutdown MaintenanceTaskType = "Shutdown"
)

// MaintenanceTaskSpec defines the desired state of MaintenanceTask
type MaintenanceTaskSpec struct {
	Type     MaintenanceTaskType  `json:"type"`
	Selector appsv1beta1.Selector `json:"selector"`
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaintenanceTask is the Schema for the maintenancetasks API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MaintenanceTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceTaskSpec   `json:"spec,omitempty"`
	Status MaintenanceTaskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaintenanceTaskList contains a list of MaintenanceTask
type MaintenanceTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaintenanceTask{}, &MaintenanceTaskList{})
}
