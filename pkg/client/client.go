package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	auth "github.com/grid-x/gxctl/pkg/auth"
)

const (
	baseURL = "https://api.ds.gridx.ai/"
)

type APIClient struct {
	Http *http.Client
}

func NewAPIClient() *APIClient {
	return &APIClient{
		Http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (apiclient *APIClient) Request(endpoint string) ([]byte, error) {
	token, err := auth.GenerateJWTToken()

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	r, err := apiclient.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
