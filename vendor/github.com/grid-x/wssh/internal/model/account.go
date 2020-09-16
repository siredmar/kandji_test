package model

import (
	"fmt"
)

// Account represents an account in our system
type Account struct {
	// UUID is the internal ID of the Account
	UUID string
	// Name
	Name string
	// Auth0Application is the internal application name within the auth0 tenant
	Auth0Application string
}

// AccountNamespaceName returns the name of the namespace for the given accountID
func AccountNamespaceName(accountID string) string {
	return fmt.Sprintf("account-%s", accountID)
}
