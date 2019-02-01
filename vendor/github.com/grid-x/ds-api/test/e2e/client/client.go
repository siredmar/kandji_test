package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
)

//APIClient to access the API via http
type APIClient struct {
	HTTP       *http.Client
	BaseURL    string
	APIVersion string
}

//Error which will be returned by the API
type Error struct {
	Message string `json:"error"`
}

//NewAPIClient to create a new instance of the APIClient
func NewAPIClient(url string) *APIClient {
	return &APIClient{
		HTTP:    &http.Client{Timeout: 10 * time.Second},
		BaseURL: url,
	}
}

//Request wraps all kind of requests
func (apiclient *APIClient) Request(endpoint string, apiversion string, method string, body []byte) (int, []byte, error) {
	switch method {
	case http.MethodGet:
		return apiclient.GetRequest(endpoint, apiversion)
	case http.MethodPost:
		return apiclient.PostRequest(endpoint, body, apiversion)
	case http.MethodPatch:
		return apiclient.PatchRequest(endpoint, body, apiversion)
	case http.MethodDelete:
		return apiclient.DeleteRequest(endpoint, apiversion)
	default:
		return -1, nil, errors.New("Just GET/POST/PATCH/DELETE are avaialable")
	}
}

//GetRequest to call via GET
func (apiclient *APIClient) GetRequest(endpoint string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodGet, nil, endpoint, apiversion)
}

//PostRequest to call via POST
func (apiclient *APIClient) PostRequest(endpoint string, body []byte, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodPost, body, endpoint, apiversion)
}

//PatchRequest to call via PATCH
func (apiclient *APIClient) PatchRequest(endpoint string, body []byte, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodPatch, body, endpoint, apiversion)
}

//DeleteRequest to call via PATCH
func (apiclient *APIClient) DeleteRequest(endpoint string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodDelete, nil, endpoint, apiversion)
}

func internalRequest(apiclient *APIClient, method string, body []byte, endpoint string, apiversion string) (int, []byte, error) {
	token, err := GenerateJWTToken()
	if err != nil {
		return -1, nil, err
	}

	url := fmt.Sprintf("%s/%s", apiclient.BaseURL, endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))

	if err != nil {
		return -1, nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", apiversion)

	r, err := apiclient.HTTP.Do(req)
	if err != nil {
		return r.StatusCode, nil, err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return r.StatusCode, nil, err
	}

	respError := Error{}
	err = json.Unmarshal(bodyBytes, &respError)

	if err != nil {
		s := fmt.Sprintf("Unexpected to unmarshal '%s'", string(bodyBytes))
		return r.StatusCode, nil, errors.New(s)
	}

	if cmp.Diff(Error{}, respError) != "" {
		return r.StatusCode, nil, errors.New(respError.Message)
	}

	return r.StatusCode, bodyBytes, nil
}
