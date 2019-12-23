package v20191210

import (
	types "github.com/grid-x/ds-api-types"
	v20191210Deployment "github.com/grid-x/ds-api-types/management/2019-12-10/deployments"
)

// DockerConfig as exposed by this API version
type DockerConfig struct {
	Metadata types.Metadata     `json:"metadata,omitempty"`
	Spec     DockerConfigSpec   `json:"spec"`
	Status   DockerConfigStatus `json:"status"`
}

// DockerConfigSpec defines the desired state of DockerConfig
type DockerConfigSpec struct {
	// Registry endpoint, eg. https://590777358184.dkr.ecr.eu-central-1.amazonaws.com
	Registry string `json:"registry"`
	// Repositories allowed to work with, eg. ["gridx/client", "gridx/supervisor-agent"]
	AllowedRepositories []string                     `json:"allowedRepositories"`
	Credentials         DockerConfigCredentails      `json:"credentials"`
	Selector            v20191210Deployment.Selector `json:"selector"`
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

// DockerConfigStatus defines the observed state of DockerConfig
type DockerConfigStatus struct {
}
