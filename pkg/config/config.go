package config

import (
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/grid-x/gxctl/pkg/client"
)

// Config contains all top-level configurables
type Config struct {
	Auth       client.AuthConfig
	ConfigFile string
	Profile    string
	UseStaging bool
}

// Read all configurables
func (c *Config) Read() error {
	return c.readAuth()
}

func (c *Config) readAuth() error {
	if c.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(c.ConfigFile)
	} else {
		xdgConfigHome := path.Join("$HOME", ".config")
		if envXdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); envXdgConfigHome != "" {
			xdgConfigHome = envXdgConfigHome
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(path.Join("$HOME", ".gxctl"))
		viper.AddConfigPath(path.Join(xdgConfigHome, "gxctl"))
		viper.AddConfigPath(".") // optionally look for config in the working directory
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&c.Auth); err != nil {
		return err
	}

	return nil
}
