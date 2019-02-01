package encoding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grid-x/ds-api/pkg/errors"
)

// IsRequest represents a request that can be decoded
type IsRequest interface {
	Validate() error
	ReadJSON(io.Reader) error
}

// IsResponse represents a response that can be encoded
type IsResponse interface {
	WriteText(io.Writer) error
	WriteJSON(io.Writer) error
}

// Response represents a response from a HTTP handler
type Response struct {
	Header  http.Header
	Status  int
	Payload IsResponse
}

// UnmarshalRequest unmarshals the request body to a given IsRequest
func UnmarshalRequest(body IsRequest, r *http.Request) error {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	// content type header should be something like "application/json..."
	if !strings.Contains(contentType, "json") {
		return errors.E(
			errors.BadRequest,
			"expected content type 'json'",
		)
	}

	err := body.ReadJSON(r.Body)
	if err != nil {
		return errors.E(
			errors.BadRequest,
			"cannot decode json payload",
		)
	}
	return r.Body.Close()
}

// Write writes a *Response to a response writer
func Write(w http.ResponseWriter, resp *Response) {
	for key, values := range resp.Header {
		w.Header().Set(key, values[0])
	}
	if resp.Status != 0 {
		w.WriteHeader(resp.Status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := resp.Payload.WriteJSON(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func kindToStatus(k errors.Kind) int {
	switch k {
	case errors.Permission:
		return http.StatusUnauthorized
	case errors.Validation:
		return http.StatusBadRequest
	case errors.NotExists:
		return http.StatusNotFound
	case errors.Service:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// ErrorPayload can be used to return an error as the payload in an response
type ErrorPayload struct {
	Error string `json:"error"`
}

// WriteText writes a text representation of ErrorPayload to the given writer
func (ep *ErrorPayload) WriteText(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%#v", ep)
	return err
}

// WriteJSON writes a JSON representation of ErrorPayload to the given writer
func (ep *ErrorPayload) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(ep)
}

// ResponseFromError creates a response from a given error
func ResponseFromError(err error) *Response {
	var status int
	var msg string
	switch e := err.(type) {
	case *errors.Error:
		status = kindToStatus(e.Kind)
		msg = e.Error()
	case error:
		status = http.StatusBadRequest
		msg = e.Error()
	default:
		status = http.StatusInternalServerError
		msg = "internal server error"
	}
	return &Response{
		Status: status,
		Payload: &ErrorPayload{
			Error: msg,
		},
	}
}
