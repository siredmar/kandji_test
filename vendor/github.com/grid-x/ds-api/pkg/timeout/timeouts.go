package timeout

import (
	"time"
)

// DurationRepository manages different timeouts
type DurationRepository struct {
}

// NewDurationRepository creates a new Duration repository containing common timeouts
func NewDurationRepository() (*DurationRepository, error) {
	r := &DurationRepository{}
	return r, nil
}

// GetSSHInactiveTimeout defines the timeout for inactive ssh connections
func (m *DurationRepository) GetSSHInactiveTimeout() time.Duration {
	return 1 * time.Hour
}

// GetSSHInactiveTicker defines the intervall of checking for inactive ssh connections
func (m *DurationRepository) GetSSHInactiveTicker() time.Duration {
	return 60 * time.Second
}
