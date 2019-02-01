package model

import (
	"time"
)

// User represents a user of our system
type User struct {
	// UUID is the internal ID of the user
	UUID string `db:"uuid"`
	// CreatedAt is the time at which the user was created
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the time at which the user was last updated
	UpdatedAt time.Time `db:"updated_at"`
	// DeletedAt is the time at which the user was deleted
	DeletedAt *time.Time `db:"deleted_at"`
	// AccountID is the ID of the Account the user belongs to
	AccountID *string `db:"account_id"`
	// FirstName is the first name of the user
	FirstName *string `db:"first_name"`
	// LastName is the last name of the user
	LastName *string `db:"last_name"`
	// Email is the email of the user
	Email string `db:"email"`
	// Auth0ID is the ID the user go from Auth0
	Auth0ID *string `db:"auth0_id"`
}
