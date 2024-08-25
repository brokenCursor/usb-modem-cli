package drivers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/brokenCursor/usb-modem-cli/logging"
	"github.com/spf13/viper"
)

// Modem interfaces

type (
	BaseModem interface {
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
		ReadAllSMS() ([]SMS, error)
	}
)

type (
	SMS struct {
		Time    time.Time
		Sender  string
		Message string
	}

	// Link statuses
	LinkStatus struct {
		State int8
	}
)

var (
	Drivers map[string]func(config *viper.Viper, logger *slog.Logger) (BaseModem, error) = map[string]func(config *viper.Viper, logger *slog.Logger) (BaseModem, error){}
	logger  *slog.Logger
)

func init() {
	logger = logging.GetGeneralLogger()
}

func IsRegistered(name string) bool {
	_, ok := Drivers[name]
	return ok
}

func RegisterDriver(name string, generator func(config *viper.Viper, logger *slog.Logger) (BaseModem, error)) {
	// Check if driver has already been registered
	if IsRegistered(name) {
		panic(fmt.Sprintf("attempted to register %s twice", name))
	}

	// Register the driver
	Drivers[name] = generator
	logger.With("name", name).Debug("driver registered")
}

func GetModemDriver(model string, config *viper.Viper, logger *slog.Logger) (BaseModem, error) {
	if !IsRegistered(model) {
		return nil, ErrUnknownModel
	}

	logger.Debug("driver instance created", "driver", model)
	return Drivers[model](config, logger)
}

func GetAvailableDrivers() *[]string {
	keys := make([]string, len(Drivers))

	i := 0
	for k := range Drivers {
		keys[i] = k
		i++
	}

	return &keys
}
