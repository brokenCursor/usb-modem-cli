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
		State int8 // 0 - down 1 - disconnecting 2 - connecting 3 - up
	}
)

var (
	driverStore map[string]func(config *viper.Viper, logger *slog.Logger) (BaseModem, error) = map[string]func(config *viper.Viper, logger *slog.Logger) (BaseModem, error){}
	logger      *slog.Logger
)

func init() {
	logger = logging.GetGeneralLogger()
}

func IsRegistered(name string) bool {
	_, ok := driverStore[name]
	return ok
}

func RegisterDriver(name string, generator func(config *viper.Viper, logger *slog.Logger) (BaseModem, error)) {
	// Check if driver has already been registered
	if IsRegistered(name) {
		panic(fmt.Sprintf("attempted to register %s twice", name))
	}

	// Register the driver
	driverStore[name] = generator
	logger.With("name", name).Debug("driver registered")
}

func GetModemDriver(name string, config *viper.Viper, logger *slog.Logger) (BaseModem, error) {
	if !IsRegistered(name) {
		return nil, ErrUnknownModel
	}

	logger.Debug("driver instance created", "driver", name)
	return driverStore[name](config, logger)
}

func GetAvailableDrivers() []string {
	keys := make([]string, len(driverStore))

	i := 0
	for k := range driverStore {
		keys[i] = k
		i++
	}

	return keys
}
