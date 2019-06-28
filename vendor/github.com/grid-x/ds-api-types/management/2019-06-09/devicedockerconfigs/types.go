package v20190609

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "github.com/grid-x/ds-api-types"
)

// DeviceDockerConfig is the Schema for the devicedockerconfigs API
type DeviceDockerConfig struct {
	Metadata types.Metadata           `json:"metadata,omitempty"`
	Spec     DeviceDockerConfigSpec   `json:"spec,omitempty"`
	Status   DeviceDockerConfigStatus `json:"status,omitempty"`
}

// DeviceDockerConfigSpec defines the desired state of DeviceDockerConfig
type DeviceDockerConfigSpec struct {
	// DeviceID is the ID of the device this config should applied to
	DeviceID string `json:"deviceID"`
	// Registry endpoint, eg. https://590777358184.dkr.ecr.eu-central-1.amazonaws.com
	Registry string `json:"registry"`
	// Repositories allowed to work with, eg. ["gridx/client", "gridx/supervisor-agent"]
	AllowedRepositories []string                `json:"allowedRepositories"`
	Credentials         DockerConfigCredentails `json:"credentials"`
}

// DockerConfigCredentails are containing credentials of different providers
type DockerConfigCredentails struct {
	AWS       *AWSCredentialProvider       `json:"aws,omitempty"`
	DockerHub *DockerHubCredentialProvider `json:"dockerhub,omitempty"`
}

// AWSCredentialProvider contains credentials required to authenticate with AWS ECR
type AWSCredentialProvider struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

// DockerHubCredentialProvider contains credentials required to authenticate with DockerHub
type DockerHubCredentialProvider struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DeviceDockerConfigStatus defines the observed state of DeviceDockerConfig
type DeviceDockerConfigStatus struct {
	AppliedAt *metav1.Time `json:"appliedAt,omitempty"`
}
