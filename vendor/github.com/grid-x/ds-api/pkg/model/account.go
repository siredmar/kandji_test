package model

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Account represents an account in our system
type Account struct {
	// UUID is the internal ID of the Account
	UUID string `db:"uuid"`
	// CreatedAt is the time at which the account was created
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the time at which the account was last updated
	UpdatedAt time.Time `db:"updated_at"`
	// DeletedAt is the time at which the account was deleted
	DeletedAt *time.Time `db:"deleted_at"`
	// Name is the human readable name of the account, e.g. gridx
	Name string `db:"account_name"`
}

// AccountNamespaceName returns the name of the namespace for the given accountID
func AccountNamespaceName(accountID string) string {
	return fmt.Sprintf("account-%s", accountID)
}

// AccountNamespace returns the namespace object of an account
func AccountNamespace(a *Account) *v1.Namespace {

	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: AccountNamespaceName(a.UUID),
		},
	}
}
