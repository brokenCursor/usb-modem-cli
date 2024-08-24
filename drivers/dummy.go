package drivers

import (
	"log/slog"

	"github.com/spf13/viper"
)

// DO NOT USE DIRECTLY
type (
	dummy struct {
		config *viper.Viper
		logger *slog.Logger
	}
)

func init() {
	RegisterDriver("dummy", newDummy)
}

func newDummy(config *viper.Viper, logger *slog.Logger) (BaseModem, error) {
	logger.With("driver", "dummy").Debug("dummy driver registered")
	return &dummy{config: config, logger: logger}, nil
}

func (m *dummy) GetModel() string {
	return "Dummy"
}
