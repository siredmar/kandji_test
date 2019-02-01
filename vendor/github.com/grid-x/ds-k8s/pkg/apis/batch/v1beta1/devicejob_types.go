/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

// DeviceJobSpec defines the desired state of DeviceJob
type DeviceJobSpec struct {
	// Selector defines which boxes to run this job on
	Selector appsv1beta1.Selector `json:"selector"`
	// Template defines the configuration of the pod
	Template corev1beta1.PodConfig `json:"template"`
}

// DeviceJobStatus defines the observed state of DeviceJob
type DeviceJobStatus struct {
	// Start time of the job will be set as soon as the scheduler handles it
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Conditions of this job
	// +optional
	Conditions []JobCondition `json:"conditions,omitempty"`
	// The statuses of the individual sub jobs
	// +optional
	SubStatuses []SubStatus `json:"statuses,omitempty"`
	// Total number of active sub jobs
	// +optional
	Active *int32 `json:"active,omitempty"`
	// Total number of succeeded sub jobs
	// +optional
	Succeeded *int32 `json:"succeeded,omitempty"`
	// Total number of failed sub jobs
	// +optional
	Failed *int32 `json:"failed,omitempty"`
}

// SubStatus describes the job of a subjob
type SubStatus struct {
	// DeviceID is the ID of the
	DeviceID string `json:"deviceID"`
	// The ID of the pod
	PodID string `json:"podID"`
	// The last time the status was synced
	LastSyncTime metav1.Time `json:"lastSyncTime,omitempty"`
	// The type of the status
	StatusType SubStatusType `json:"type"`
}

// SubStatusType describes the type of the sub status
type SubStatusType string

const (
	// SubStatusTypePending indicates that the corresponding job/pod
	// hasn't started
	SubStatusTypePending SubStatusType = "Pending"
	// SubStatusTypeRunning indicates that the corresponding job/pod
	// is running
	SubStatusTypeRunning SubStatusType = "Running"
	// SubStatusTypeSucceeded indicates that the corresponding job/pod has
	// succeeded
	SubStatusTypeSucceeded SubStatusType = "Succeeded"
	// SubStatusTypeFailed indicates that the corresponding job/pod has
	// failed
	SubStatusTypeFailed SubStatusType = "Failed"
)

// JobConditionType indicates the type of the job condition
type JobConditionType string

// These are valid conditions of a job.
const (
	// JobComplete means the job has completed its execution.
	JobComplete JobConditionType = "Complete"
	// JobFailed means the job has failed its execution.
	JobFailed JobConditionType = "Failed"
)

// JobCondition is a condition of a job
type JobCondition struct {
	// Type of job condition, Complete or Failed.
	Type JobConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1beta1.ConditionStatus `json:"status"`
	// Last time the condition was checked.
	// +optional
	LastProbeTime *metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceJob is the Schema for the devicejobs API
// +k8s:openapi-gen=true
type DeviceJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceJobSpec   `json:"spec,omitempty"`
	Status DeviceJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceJobList contains a list of DeviceJob
type DeviceJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceJob{}, &DeviceJobList{})
}
