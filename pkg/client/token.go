package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Token struct {
	Exp int `json:"exp"`
}

func TokenValid(t string) error {
	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		return fmt.Errorf("Invalid Token")
	}

	parsedToken := Token{}

	if l := len(parts[1]) % 4; l > 0 {
		parts[1] += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("Invalid Token")
	}

	if err := json.Unmarshal(decoded, &parsedToken); err != nil {
		return fmt.Errorf("Invalid Token")
	}

	if time.Now().After(time.Unix(int64(parsedToken.Exp), 0)) {
		return fmt.Errorf("Token expired")
	}

	return nil
}
