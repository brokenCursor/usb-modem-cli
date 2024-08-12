package common

import (
	"fmt"
	"log/slog"

	"github.com/brokenCursor/usb-modem-cli/config"
	"github.com/brokenCursor/usb-modem-cli/logging"
	"github.com/spf13/viper"
)

// Modem interfaces

type (
	BaseModem interface {
		SetHost(host string) error
		GetHost() string
		GetModel() string
	}

	ModemCell interface {
		BaseModem

		GetCellConnStatus() (*LinkStatus, error)
		ConnectCell() error
		DisconnectCell() error
	}

	ModemSMS interface {
		BaseModem

		SendSMS(phone string, message string) error
		// GetAllSMS() ([]SMS, error)
	}
)

type (
	SMS struct {
		Sender  string
		Message string
	}

	// Link statuses
	LinkStatus struct {
		State int8
	}
)

var (
	drivers      map[string]func(host string) BaseModem = map[string]func(host string) BaseModem{}
	driverConfig *viper.Viper
	logger       *slog.Logger
)

func init() {
	logger = logging.GetGeneralLogger()
	driverConfig = config.Sub("driver")
}

func IsRegistered(name string) bool {
	// Check if driver has already been registered
	_, ok := drivers[name]
	return ok
}

func RegisterDriver(name string, generator func(ip string) BaseModem) (*viper.Viper, *slog.Logger) {
	// Check if driver has already been registered
	if IsRegistered(name) {
		panic(fmt.Sprintf("attempted to register %s twice", name))
	}

	// Register the driver
	drivers[name] = generator
	logger.With("name", name).Debug("driver registered")

	return driverConfig, logging.GetDriverLogger(name)
}

func GetModemDriver(model string, host string) (BaseModem, error) {
	if !IsRegistered(model) {
		return nil, ErrUnknownModel
	}

	logger.Debug("driver instance created", "driver", model, "host", host)
	return drivers[model](host), nil
}

func GetAvailableDrivers() *[]string {
	keys := make([]string, len(drivers))

	i := 0
	for k := range drivers {
		keys[i] = k
		i++
	}

	return &keys
}
