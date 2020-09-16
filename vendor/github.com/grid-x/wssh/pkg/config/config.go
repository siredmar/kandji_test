package config

// ServerConfig contains all configurable params of the server
type ServerConfig struct {
	ListenAddr string
	ExternalAddr string
	DeviceImage string
	DebugDeviceSpawn bool
	DebugDeviceAddr string
}
