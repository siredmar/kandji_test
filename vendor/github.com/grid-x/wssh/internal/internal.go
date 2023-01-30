package internal

import (
	"context"
	"time"
)

const (
	// DevicePodKey for usage with context value
	DevicePodKey = "DevicePod"
	// SessionContextKey for usage with context value
	SessionContextKey = "SessionContext"
)

// SessionContext contains the spec for launching a Session
type SessionContext struct {
	SessionID string
	AccountID string
	AgentID string
	DeviceID string
}

// DefaultDeviceImage is set via Makefile build
var DefaultDeviceImage string

// NewValueOnlyContext returns a new ValueOnlyContext
func NewValueOnlyContext(ctx context.Context) context.Context {
	return ValueOnlyContext{ctx}
}
// ValueOnlyContext is a Context containing only values of another Context
// https://stackoverflow.com/a/59348871
type ValueOnlyContext struct {
	context.Context
}
// Deadline returns nil
func (ValueOnlyContext) Deadline() (deadline time.Time, ok bool) {
	return
}
// Done returns nil
func (ValueOnlyContext) Done() <-chan struct{} {
	return nil
}
// Err returns nil
func (ValueOnlyContext) Err() error {
	return nil
}
