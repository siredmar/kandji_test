package v20200504

// LogMessage represents a single log message
type LogMessage struct {
	// Timestamp of the log message specified in nanoseconds
	Timestamp string `json:"ts"`
	// The severity field should range from 0 to 6, and identifies the importance of this event, using the classic scale "finest, finer, fine, info, warning, error, fatal"
	Severity int `json:"sev"`
	// Log message
	Message string `json:"msg"`
	// The attributes field specifies the fields of the event and can contain arbitrary objects
	Attributes map[string]interface{} `json:"attr,omitempty"`
}
