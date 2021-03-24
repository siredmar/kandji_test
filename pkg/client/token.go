package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Token used for DS API authentication
type Token string

func (t Token) String() string {
	return string(t)
}

// MarshalJSON Token
func (t Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Validate Token
func (t *Token) Validate() error {
	token, err := t.parse()

	if err != nil {
		return fmt.Errorf("invalid token")
	}

	if time.Now().After(time.Unix(int64(token.Exp), 0)) {
		return fmt.Errorf("Token expired")
	}

	return nil
}

type tokenParsed struct {
	Exp           int    `json:"exp"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Iat           int    `json:"iat"`
	Nonce         string `json:"nonce"`
}

// Email of this Token
func (t *Token) Email() string {
	token, err := t.parse()
	if err != nil {
		return ""
	}
	return token.Email
}

func (t *Token) parse() (*tokenParsed, error) {
	parts := strings.Split(t.String(), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("parse")
	}

	if l := len(parts[1]) % 4; l > 0 {
		parts[1] += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode base64")
	}

	token := &tokenParsed{}
	if err := json.Unmarshal(decoded, token); err != nil {
		return nil, fmt.Errorf("unmarshal JSON")
	}

	return token, nil
}
