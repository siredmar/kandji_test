package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"

	api "github.com/grid-x/gxctl/pkg/api"
	auth "github.com/grid-x/gxctl/pkg/auth"
	errors "github.com/grid-x/gxctl/pkg/error"
)

const (
	baseURL = "https://api.ds.gridx.ai"
	//baseURL = "http://127.0.0.1:8080"
)

type APIClient struct {
	Http *http.Client
}

type Error struct {
	Message string `json:"error"`
}

func NewAPIClient() *APIClient {
	return &APIClient{
		Http: &http.Client{Timeout: 10 * time.Second},
	}
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
	token, err := auth.GenerateJWTToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
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
		s := fmt.Sprintf("Unexpected to unmarshal '%s'", string(bodyBytes))
		return nil, errors.ServerError(s)
	}

	if cmp.Diff(Error{}, respError) != "" {
		return nil, errors.ServerError(respError.Message)
	}

	return bodyBytes, nil
}
