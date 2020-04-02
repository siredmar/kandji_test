package v20200318

import (
	"time"
)

// CPUMetric contains CPU related metrics
type CPUMetric struct {
	// averaged CPU usage [0, 1]
	Usage float32 `json:"usage"`
}

// RAMMetric contains memory related metrics
type RAMMetric struct {
	Available int `json:"available"` // [kB]
}

// StorageMetric contains storage related metrics
type StorageMetric struct {
	Available int `json:"available"` // [kB]
}

// SystemMetric contains all system related metrics at a specific time
type SystemMetric struct {
	ReadAt  time.Time                `json:"readAt"`
	CPU     *CPUMetric               `json:"CPU",omitempty`
	RAM     *RAMMetric               `json:"RAM,omitempty"`
	Storage map[string]StorageMetric `json:"storage,omitempty"`
}
