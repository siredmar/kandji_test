package config

import "time"

// ServerConfig contains all configurable params of the server
type ServerConfig struct {
	ListenAddr          string
	ExternalAddr        string
	PromListenAddr      string
	DSAddr              string
	KetoAddr            string
	DeviceImage         string
	DebugDeviceSpawn    bool
	DebugDeviceAddr     string
	DeviceOnlineTimeout time.Duration
	AuditBucket         string
	AuditUploadLimitMB  uint
	LogLevel            string
	CompressLogs        bool
}

// DeviceConfig contains all configurable params of the device
type DeviceConfig struct {
	ServerAddr   string
	DeviceAddr   string
	AuthTimeout  time.Duration
	CompressLogs bool
}
