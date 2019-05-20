package v20190509

// CreateRequest represents the request type
type CreateRequest struct {
	Messages []*LogMessage `json:"logs"`
}

// CreateResponse represents the response type
type CreateResponse struct{}
