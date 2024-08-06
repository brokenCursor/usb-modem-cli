package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var config *viper.Viper

func init() {
	// Get config path
	dir, err := os.UserConfigDir()
	if err != nil {
		panic("failed to get user config dir")
	}

	// Setup configuration
	config = viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")

	sep := string(os.PathSeparator)
	config.AddConfigPath(dir + sep + "modem-cli")

	// -- Defaults -- //
	config.SetDefault("modem.model", "dummy")
	config.SetDefault("modem.host", "127.0.0.1")

	config.SetDefault("logging.general", "error")
	config.SetDefault("logging.driver", "error")

	config.SetDefault("driver.cmd_ttl", 10)
	// -- Defaults -- //

	err = config.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func Sub(name string) *viper.Viper {
	return config.Sub(name)
}
