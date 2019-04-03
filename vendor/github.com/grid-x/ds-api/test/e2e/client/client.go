package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

const (
	// the header that contains the signature
	signatureHeader = "X-GRIDX"
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

// TokenRequest represents a request to get an authorization token.
type TokenRequest struct {
	PublicKey *string `json:"publicKey,omitempty"`
	IDData    *string `json:"idData,omitempty"`
}

// Token represents a device service token.
type Token struct {
	Token     *string    `json:"token,omitempty"`
	Subject   *string    `json:"subject,omitempty"`
	Issuer    *string    `json:"issuer,omitempty"`
	IssuedAt  *time.Time `json:"issuedAt,omitempty"`
	NotBefore *time.Time `json:"notBefore,omitempty"`
	ExpiredAt *time.Time `json:"expiredAt,omitempty"`
}

// String returns string representation of the underlying token.
func (t *Token) String() string {
	if t.Token == nil {
		return ""
	}
	return *t.Token
}

// NewAPIClient to create a new instance of the APIClient
func NewAPIClient(url string) *APIClient {
	return &APIClient{
		HTTP:    &http.Client{Timeout: 10 * time.Second},
		BaseURL: url,
	}
}

// Request wraps all kind of requests
func (apiclient *APIClient) Request(endpoint string, apiversion string, method string, body []byte, additionalHeader map[string]string, token string) (int, []byte, error) {
	switch method {
	case http.MethodGet:
		return apiclient.GetRequest(endpoint, additionalHeader, token, apiversion)
	case http.MethodPost:
		return apiclient.PostRequest(endpoint, body, additionalHeader, token, apiversion)
	case http.MethodPatch:
		return apiclient.PatchRequest(endpoint, body, additionalHeader, token, apiversion)
	case http.MethodDelete:
		return apiclient.DeleteRequest(endpoint, additionalHeader, token, apiversion)
	default:
		return -1, nil, errors.New("Just GET/POST/PATCH/DELETE are avaialable")
	}
}

// GetRequest to call via GET
func (apiclient *APIClient) GetRequest(endpoint string, additionalHeader map[string]string, token string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodGet, nil, additionalHeader, token, endpoint, apiversion)
}

// PostRequest to call via POST
func (apiclient *APIClient) PostRequest(endpoint string, body []byte, additionalHeader map[string]string, token string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodPost, body, additionalHeader, token, endpoint, apiversion)
}

// PatchRequest to call via PATCH
func (apiclient *APIClient) PatchRequest(endpoint string, body []byte, additionalHeader map[string]string, token string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodPatch, body, additionalHeader, token, endpoint, apiversion)
}

// DeleteRequest to call via PATCH
func (apiclient *APIClient) DeleteRequest(endpoint string, additionalHeader map[string]string, token string, apiversion string) (int, []byte, error) {
	return internalRequest(apiclient, http.MethodDelete, nil, additionalHeader, token, endpoint, apiversion)
}

// JSONEscape escapes a string (eg. multiline) properly to be send via JSON
func JSONEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// GetDeviceToken requests a token from the device services.
func (apiclient *APIClient) GetDeviceToken(privKey *rsa.PrivateKey, pubKey string, endpoint, apiversion string) (*Token, error) {
	iddata := "Test"
	tr := &TokenRequest{}
	tr.IDData = &iddata
	tr.PublicKey = &pubKey
	marshalled, err := json.Marshal(&tr)
	if err != nil {
		return nil, err
	}

	signedContent, err := Sign(marshalled, privKey)
	if err != nil {
		return nil, err
	}

	additionalHeader := map[string]string{}
	additionalHeader[signatureHeader] = string(signedContent)

	statusCode, resp, err := apiclient.PostRequest(endpoint, marshalled, additionalHeader, "", apiversion)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("Wrong statuscode: %d", statusCode)
	}

	t := new(Token)
	err = json.Unmarshal(resp, &t)

	if err := parseToken(t); err != nil {
		return nil, err
	}
	return t, nil
}

// parseToken parses the token and populates the struct with details.
func parseToken(t *Token) error {
	if t.Token == nil || *t.Token == "" {
		return fmt.Errorf("couldn't parse token: token is empty")
	}
	var claims jwt.StandardClaims

	if claims.Subject != "" {
		t.Subject = &claims.Subject
	}
	if claims.Issuer != "" {
		t.Issuer = &claims.Issuer
	}
	if claims.IssuedAt != 0 {
		ts := time.Unix(claims.IssuedAt, 0)
		t.IssuedAt = &ts
	}
	if claims.NotBefore != 0 {
		ts := time.Unix(claims.NotBefore, 0)
		t.NotBefore = &ts
	}
	if claims.ExpiresAt != 0 {
		ts := time.Unix(claims.ExpiresAt, 0)
		t.ExpiredAt = &ts
	}
	return nil
}

func internalRequest(apiclient *APIClient, method string, body []byte, additionalHeader map[string]string, token string, endpoint string, apiversion string) (int, []byte, error) {
	if token == "" {
		var err error
		token, err = GenerateJWTToken()
		if err != nil {
			return -1, nil, err
		}
	}

	url := fmt.Sprintf("%s/%s", apiclient.BaseURL, endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))

	if err != nil {
		return -1, nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", apiversion)

	for k, v := range additionalHeader {
		req.Header.Add(k, v)
	}

	r, err := apiclient.HTTP.Do(req)
	if err != nil {
		if r != nil {
			return r.StatusCode, nil, err
		}
		return -1, nil, err
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

// GetWebsocketConnection creates a new websocket connection
func (apiclient *APIClient) GetWebsocketConnection(endpoint, token, apiversion string) (*websocket.Conn, error) {
	if token == "" {
		var err error
		token, err = GenerateJWTToken()
		if err != nil {
			return nil, err
		}
	}

	h := http.Header{}
	h.Set("Origin", fmt.Sprintf("%s/%s", apiclient.BaseURL, endpoint))
	h.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	h.Set("Accept", apiversion)

	// BaseURL without http://
	url := fmt.Sprintf("ws://%s/%s", apiclient.BaseURL[7:], endpoint)

	conn, resp, err := websocket.DefaultDialer.Dial(url, h)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			fmt.Printf("handshake failed with status %d", resp.StatusCode)
		}

		return nil, err
	}

	return conn, nil
}
