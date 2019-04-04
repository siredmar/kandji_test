/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package v1beta1

import (
	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DockerConfigSpec defines the desired state of DockerConfig
type DockerConfigSpec struct {
	Registry    string                  `json:"registry"`
	Credentials DockerConfigCredentails `json:"credentials"`
	Selector    appsv1beta1.Selector    `json:"selector"`
}

// DockerConfigCredentails are containing credentials of different providers
type DockerConfigCredentails struct {
	// +optional
	AWS *AWSCredentialProvider `json:"aws,omitempty"`
	// +optional
	DockerHub *DockerHubCredentialProvider `json:"dockerhub,omitempty"`
}

// AWSCredentialProvider contains credentials required to authenticate with AWS ECR
type AWSCredentialProvider struct {
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

// DockerHubCredentialProvider contains credentials required to authenticate with DockerHub
type DockerHubCredentialProvider struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DockerConfigStatus defines the observed state of DockerConfig
type DockerConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DockerConfig is the Schema for the dockerconfigs API
// +k8s:openapi-gen=true
type DockerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerConfigSpec   `json:"spec,omitempty"`
	Status DockerConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DockerConfigList contains a list of DockerConfig
type DockerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerConfig{}, &DockerConfigList{})
}
