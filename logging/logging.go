package logging

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/brokenCursor/usb-modem-cli/config"
)

var (
	general *slog.Logger
	driver  *slog.Logger
)

func init() {
	logConfig := config.Sub("logging")
	// Setup levels
	generalLevel, err := strToLevel(logConfig.GetString("general"))
	if err != nil {
		slog.Error("logging.general: ", err.Error(), nil)
	}

	// Setup levels
	driverLevel, err := strToLevel(logConfig.GetString("driver"))
	if err != nil {
		slog.Error("logging.driver: ", err.Error(), nil)
	}

	// Setup loggers
	general = slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{
			Level:     generalLevel,
			AddSource: false,
		}))

	driver = slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{
			Level:     driverLevel,
			AddSource: true,
		}))

	general.Debug("logging setup done")
}

func strToLevel(level string) (res slog.Level, err error) {
	switch level {
	case "debug":
		res = slog.LevelDebug
	case "info":
		res = slog.LevelInfo
	case "warn":
		res = slog.LevelWarn
	case "error":
		res = slog.LevelError
	default:
		return slog.LevelInfo, fmt.Errorf("no %s logging level", level)
	}

	return res, nil
}

func GetGeneralLogger() *slog.Logger {
	return general
}

func GetDriverLogger(name string) *slog.Logger {
	return driver
}
