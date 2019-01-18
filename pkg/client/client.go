package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"net/http"
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
	auth "github.com/grid-x/gxctl/pkg/auth"
	errors "github.com/grid-x/gxctl/pkg/error"
)

const (
	baseURL = "https://api.ds.gridx.ai"
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

func (apiclient *APIClient) GetRequest(endpoint string) ([]byte, error) {
	return request(apiclient, http.MethodGet, nil, endpoint)
}

func (apiclient *APIClient) PostRequest(endpoint string, v interface{}) ([]byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return request(apiclient, http.MethodPost, body, endpoint)
}

func request(apiclient *APIClient, method string, body []byte, endpoint string) ([]byte, error) {
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
		return nil, err
	}

	if cmp.Diff(Error{}, respError) != "" {
		return nil, errors.ServerError(respError.Message)
	}

	return bodyBytes, nil
}
