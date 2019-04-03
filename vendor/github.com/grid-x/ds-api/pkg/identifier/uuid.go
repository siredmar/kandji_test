package identifier

import (
	"github.com/google/uuid"
)

// GetUUID represents a function returning a v4 uuid
type GetUUID func() uuid.UUID

// UUIDRepository provides methods to manage connections
type UUIDRepository struct {
	uuid GetUUID
}

// NewUUIDRepository creates a new connection repository with the given settings
func NewUUIDRepository(
	uuid GetUUID,
) (*UUIDRepository, error) {
	r := &UUIDRepository{
		uuid: uuid,
	}

	return r, nil
}

// Get retrieves a random v4 uuid
func (r *UUIDRepository) Get() uuid.UUID {
	return r.uuid()
}
