package v20200318

// CreateRequest represents the request type
type CreateRequest struct {
	SystemMetrics []SystemMetric `json:"systemMetrics"`
}

// CreateResponse represents the response type
type CreateResponse struct{}
