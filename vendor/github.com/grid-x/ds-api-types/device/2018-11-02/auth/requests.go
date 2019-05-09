package v20181102

// GetTokenRequest represents the request type
type GetTokenRequest struct {
	PublicKey string `json:"publicKey"`
	IDData    string `json:"idData"`
}

// GetTokenResponse represents the response type
type GetTokenResponse struct {
	Token string `json:"token"`
}
