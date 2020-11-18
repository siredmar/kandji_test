package config

import "time"

// ServerConfig contains all configurable params of the server
type ServerConfig struct {
	ListenAddr          string
	ExternalAddr        string
	DSAddr              string
	DeviceImage         string
	DebugDeviceSpawn    bool
	DebugDeviceAddr     string
	DeviceOnlineTimeout time.Duration
}
