package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/grid-x/gxctl/internal/version"
	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

const (
	baseURL        = "api.ds.gridx.ai"
	baseURLStaging = "api.staging.ds.gridx.ai"
)

type AuthConfig struct {
	Profiles []struct {
		Name    string `yaml:"name"`
		Default bool   `yaml:"default,omitempty"`
		Staging bool   `yaml:"staging,omitempty"`
		Auth    struct {
			Auth0Tenant   string `yaml:"auth0Tenant"`
			Auth0ClientID string `yaml:"auth0ClientID"`
			Token         Token  `yaml:"token"`
		} `yaml:"auth"`
	} `yaml:"profiles"`
}

type APIClient struct {
	Http    *http.Client
	limiter *rate.Limiter
	Auth    *AuthConfig
	Profile *string
}

type Error struct {
	Error struct {
		Message string `json:"message"`
	} `json:"Error"`
}

type RequestResult struct {
	Profile string
	Body    []byte
	Err     error
}

func NewAPIClient(auth *AuthConfig, profile string) *APIClient {
	return &APIClient{
		Http:    &http.Client{Timeout: 60 * time.Second},
		limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		Auth:    auth,
		Profile: &profile,
	}
}

// GetToken of current profile
func (apiclient *APIClient) GetToken() (*Token, error) {
	p, err := apiclient.resolveProfileNames()
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			"resolve profile",
		)
	}
	if len(p) > 1 {
		return nil, errors.E(
			errors.Internal,
			"more than 1 profile",
		)
	}

	token, err := apiclient.GetTokenFromAuthConfig(p[0])
	return &token, err
}

//GetRequest to call via GET
func (apiclient *APIClient) GetRequest(endpoint string) ([]byte, error) {
	result, err := apiclient.internalRequest(http.MethodGet, nil, endpoint)
	if err != nil {
		return nil, err
	}
	return result[0].Body, result[0].Err
}

//GetMultiRequest to call via GET
func (apiclient *APIClient) GetMultiRequest(endpoint string) ([]RequestResult, error) {
	result, err := apiclient.internalRequest(http.MethodGet, nil, endpoint)
	if err != nil {
		return nil, err
	}
	return result, err
}

//PostRequest to call via POST
func (apiclient *APIClient) PostRequest(endpoint string, v interface{}) ([]byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	result, err := apiclient.internalRequest(http.MethodPost, body, endpoint)
	if err != nil {
		return nil, err
	}
	return result[0].Body, result[0].Err
}

//PatchRequest to call via PATCH
func (apiclient *APIClient) PatchRequest(endpoint string, v api.Resource, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", endpoint, id)
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	result, err := apiclient.internalRequest(http.MethodPatch, body, url)
	if err != nil {
		return nil, err
	}
	return result[0].Body, result[0].Err
}

//DeleteRequest to call via DELETE
func (apiclient *APIClient) DeleteRequest(endpoint string, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", endpoint, id)
	result, err := apiclient.internalRequest(http.MethodDelete, nil, url)
	if err != nil {
		return nil, err
	}
	return result[0].Body, result[0].Err
}

func (apiclient *APIClient) internalRequest(method string, body []byte, endpoint string) ([]RequestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := apiclient.limiter.Wait(ctx); err != nil {
		cancel()
		return nil, errors.E(
			errors.Internal,
			"Exceed timeout while haning in rate-limit",
		)
	}
	cancel()

	profileNames, err := apiclient.resolveProfileNames()
	if err != nil {
		return nil, errors.E(errors.Invalid, "resolve profile", err)
	}

	if len(profileNames) > 1 && method != http.MethodGet {
		return nil, errors.E(errors.Invalid, "more than 1 match on non-GET method")
	}

	results := make([]RequestResult, len(profileNames))

	for i, p := range profileNames {
		results[i] = RequestResult{
			Profile: p,
		}
		result := &results[i]

		token, err := apiclient.GetTokenFromAuthConfig(p)
		if err != nil {
			result.Err = errors.E(errors.Invalid, "Token not found", err)
			continue
		}

		if err := token.Validate(); err != nil {
			result.Err = err
			continue
		}

		base := baseURL
		if isStaging := apiclient.IsStaging(p); isStaging {
			base = baseURLStaging
		}

		url := fmt.Sprintf("https://%s/%s", base, endpoint)
		req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			result.Err = err
			continue
		}

		req.Header.Add("User-Agent", version.UserAgent())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", api.APIVersion)

		r, err := apiclient.Http.Do(req)
		if err != nil {
			result.Err = err
			continue
		}
		defer r.Body.Close()

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			result.Err = err
			continue
		}

		if r.StatusCode == http.StatusUnauthorized {
			result.Err = errors.E(
				errors.Permission,
				"Invalid access token",
				"Invalid access token",
			)
			continue
		}

		respError := Error{}
		err = json.Unmarshal(bodyBytes, &respError)
		if err != nil {
			// non-JSON (text) response
			if r.StatusCode == http.StatusNotFound {
				result.Err = errors.E(
					errors.NotExists,
				)
				continue
			}
			result.Err = errors.E(
				errors.Internal,
				"Unknown server error",
			)
			continue
		}

		errorMsg := respError.Error.Message
		if errorMsg == "" {
			errorMsg = string(bodyBytes)
		}

		if r.StatusCode == http.StatusNotFound {
			result.Err = errors.E(
				errors.NotExists,
				errorMsg,
			)
			continue
		}

		if r.StatusCode == http.StatusConflict {
			result.Err = errors.E(
				errors.Exist,
				errorMsg,
			)
			continue
		}

		if r.StatusCode >= 400 {
			result.Err = errors.E(
				errors.Internal,
				errorMsg,
			)
			continue
		}
		result.Body = bodyBytes

	}
	return results, nil
}

func (apiclient *APIClient) IsStaging(profileName string) bool {
	var isStaging bool

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == profileName {
			isStaging = profile.Staging
		}
	}

	return isStaging
}

func (apiclient *APIClient) GetTokenFromAuthConfig(profileName string) (Token, error) {
	var token Token

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == profileName {
			token = profile.Auth.Token
		}
	}

	if token == "" {
		return token, fmt.Errorf("Token for profile %s not configured", profileName)
	}

	return token, nil
}

func (apiclient *APIClient) SetTokenInAuthConfig(token Token) error {
	p, err := apiclient.resolveProfileNames()
	if err != nil {
		return err
	}
	if len(p) > 1 {
		return errors.E(errors.Invalid, "more than 1 match")
	}

	for i, profile := range apiclient.Auth.Profiles {
		if profile.Name == p[0] {
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

	p, err := apiclient.resolveProfileNames()
	if err != nil {
		return "", err
	}
	if len(p) > 1 {
		return "", errors.E(errors.Invalid, "more than 1 match")
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == p[0] {
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

	p, err := apiclient.resolveProfileNames()
	if err != nil {
		return "", err
	}
	if len(p) > 1 {
		return "", errors.E(errors.Invalid, "more than 1 match")
	}

	for _, profile := range apiclient.Auth.Profiles {
		if profile.Name == p[0] {
			clientID = profile.Auth.Auth0ClientID
		}
	}

	if clientID == "" {
		return clientID, fmt.Errorf("Auth0 clientID for profile %s not configured", p)
	}

	return clientID, nil
}

func (apiclient *APIClient) resolveProfileNames() ([]string, error) {
	if len(apiclient.Auth.Profiles) == 0 {
		return nil, fmt.Errorf("No profiles found. Please add a profile to $HOME/.gxctl/config.yaml")
	}

	if *apiclient.Profile == "*" {
		var profileIDs []string
		for _, p := range apiclient.Auth.Profiles {
			profileIDs = append(profileIDs, p.Name)
		}
		return profileIDs, nil
	}

	var defaultProfile string
	for _, profile := range apiclient.Auth.Profiles {
		if profile.Default {
			defaultProfile = profile.Name
		}
	}

	if *apiclient.Profile == "" {
		if defaultProfile == "" {
			return nil, fmt.Errorf("No default profile configured. Use --profile or configure a default profile")
		} else {
			return []string{defaultProfile}, nil
		}
	}

	for _, p := range apiclient.Auth.Profiles {
		if p.Name == *apiclient.Profile {
			return []string{*apiclient.Profile}, nil
		}
	}

	return nil, fmt.Errorf("Profile %s not found", *apiclient.Profile)
}
