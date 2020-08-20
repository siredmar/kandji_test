package internal

import (
	"context"
	"time"
)

const (
	// AccountID for usage with context
	AccountID = "AccountID"
	// DeviceID  for usage with context
	DeviceID = "DeviceID"
	// DevicePod for usage with context
	DevicePod = "DevicePod"
)

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
