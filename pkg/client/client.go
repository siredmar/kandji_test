package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	api "github.com/grid-x/gxctl/pkg/api"
	errors "github.com/grid-x/gxctl/pkg/error"
)

const (
	baseURL        = "api.ds.gridx.ai"
	baseURLStaging = "api.staging.ds.gridx.ai"
)

type AuthConfig struct {
	Profiles []struct {
		Name    string `yaml:"name"`
		Default bool   `yaml:"default,omitempty"`
		Auth    struct {
			Auth0Tenant   string `yaml:"auth0Tenant"`
			Auth0ClientID string `yaml:"auth0ClientID"`
			Token         string `yaml:"token"`
		} `yaml:"auth"`
	} `yaml:"profiles"`
}

type APIClient struct {
	Http    *http.Client
	Staging *bool
	Auth    *AuthConfig
	Profile *string
}

type Error struct {
	Error struct {
		Message string `json:"message"`
	} `json:"Error"`
}

func NewAPIClient(staging *bool, auth *AuthConfig, profile *string) *APIClient {
	return &APIClient{
		Http:    &http.Client{Timeout: 10 * time.Second},
		Auth:    auth,
		Staging: staging,
		Profile: profile,
	}
}

//GetWebsocketConnection returns a websocket connection
func (apiclient *APIClient) GetWebsocketConnection(endpoint string, additionalHeaders map[string]string) (*websocket.Conn, error) {
	token, err := apiclient.getTokenFromAuthConfig()
	if err != nil {
		return nil, err
	}

	base := baseURL
	if *apiclient.Staging {
		base = baseURLStaging
	}

	h := http.Header{}
	h.Set("Origin", fmt.Sprintf("https://%s/%s", baseURL, endpoint))
	h.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.Set("Accept", api.APIVersion)
	for k, v := range additionalHeaders {
		h.Set(k, v)
	}

	url := fmt.Sprintf("wss://%s/%s", base, endpoint)

	conn, resp, err := websocket.DefaultDialer.Dial(url, h)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			fmt.Printf("handshake failed with status %d", resp.StatusCode)
		}

		return nil, err
	}

	return conn, nil
}

//GetRequest to call via GET
func (apiclient *APIClient) GetRequest(endpoint string) ([]byte, error) {
	return internalRequest(apiclient, http.MethodGet, nil, endpoint)
}

//PostRequest to call via POST
func (apiclient *APIClient) PostRequest(endpoint string, v interface{}) ([]byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return internalRequest(apiclient, http.MethodPost, body, endpoint)
}

//PatchRequest to call via PATCH
func (apiclient *APIClient) PatchRequest(endpoint string, v interface{}, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", endpoint, id)
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return internalRequest(apiclient, http.MethodPatch, body, url)
}

//DeleteRequest to call via DELETE
func (apiclient *APIClient) DeleteRequest(endpoint string, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", endpoint, id)
	return internalRequest(apiclient, http.MethodDelete, nil, url)
}

func internalRequest(apiclient *APIClient, method string, body []byte, endpoint string) ([]byte, error) {
	token, err := apiclient.getTokenFromAuthConfig()
	if err != nil {
		return nil, err
	}

	base := baseURL
	if *apiclient.Staging {
		base = baseURLStaging
	}

	url := fmt.Sprintf("https://%s/%s", base, endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", api.APIVersion)

	r, err := apiclient.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if r.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Invalid access token: %s", string(bodyBytes))
	}

	respError := Error{}
	err = json.Unmarshal(bodyBytes, &respError)
	if err != nil {
		return nil, err
	}

	if respError.Error.Message != "" {
		return nil, errors.ServerError(respError.Error.Message)
	}

	return bodyBytes, nil
}

func (apiclient *APIClient) getTokenFromAuthConfig() (string, error) {
	var token string

	p, err := apiclient.resolveProfile()
	if err != nil {
		return "", err
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == p {
			token = profile.Auth.Token
		}
	}

	if token == "" {
		return token, fmt.Errorf("Token for profile %s not configured", p)
	}

	return token, nil
}

func (apiclient *APIClient) SetTokenInAuthConfig(token string) error {
	p, err := apiclient.resolveProfile()
	if err != nil {
		return err
	}

	for i, profile := range apiclient.Auth.Profiles {
		if profile.Name == p {
			apiclient.Auth.Profiles[i].Auth.Token = token
		}
	}

	c, err := yaml.Marshal(apiclient.Auth)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(viper.ConfigFileUsed(), c, 0644)
}

func (apiclient *APIClient) GetAuth0TenantFromAuthConfig() (string, error) {
	var tenant string

	p, err := apiclient.resolveProfile()
	if err != nil {
		return "", err
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == p {
			tenant = profile.Auth.Auth0Tenant
		}
	}

	if tenant == "" {
		return tenant, fmt.Errorf("Auth0 tenant for profile %s not configured", p)
	}

	return tenant, nil
}

func (apiclient *APIClient) GetAuth0ClientIDFromAuthConfig() (string, error) {
	var clientID string

	p, err := apiclient.resolveProfile()
	if err != nil {
		return "", err
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == p {
			clientID = profile.Auth.Auth0ClientID
		}
	}

	if clientID == "" {
		return clientID, fmt.Errorf("Auth0 clientID for profile %s not configured", p)
	}

	return clientID, nil
}

func (apiclient *APIClient) resolveProfile() (string, error) {
	if len(apiclient.Auth.Profiles) == 0 {
		return "", fmt.Errorf("No profiles found. Please add a profile to $HOME/.gxctl/config.yaml")
	}

	var defaultProfile string
	for _, profile := range apiclient.Auth.Profiles {
		if profile.Default {
			defaultProfile = profile.Name
		}
	}

	if *apiclient.Profile == "" {
		if defaultProfile == "" {
			return "", fmt.Errorf("No default profile configured. Use --profile or configure a default profile")
		} else {
			return defaultProfile, nil
		}
	}

	for _, p := range apiclient.Auth.Profiles {
		if p.Name == *apiclient.Profile {
			return *apiclient.Profile, nil
		}
	}

	return "", fmt.Errorf("Profile %s not found", *apiclient.Profile)
}
