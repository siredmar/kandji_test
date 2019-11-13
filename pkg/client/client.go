package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

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
			Token string `yaml:"token"`
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

	if len(apiclient.Auth.Profiles) == 0 {
		return token, fmt.Errorf("No profiles found. Please add a profile to $HOME/.gxctl/config.yaml")
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Default {
			token = profile.Auth.Token
		}
	}

	if token == "" && *apiclient.Profile == "" {
		return token, fmt.Errorf("No default profile configured. Use --profile")
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == *apiclient.Profile {
			token = profile.Auth.Token
		}
	}

	if token == "" {
		return token, fmt.Errorf("Profil %s not found", *apiclient.Profile)
	}

	return token, nil
}
