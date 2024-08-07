package drivers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/brokenCursor/usb-modem-cli/config"
	"github.com/brokenCursor/usb-modem-cli/logging"
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
	drivers    map[string]func(host string) BaseModem = map[string]func(host string) BaseModem{}
	httpClient *http.Client

	logger  *slog.Logger
	dLogger *slog.Logger
)

func init() {
	logger = logging.GetGeneralLogger()
	// TODO: fix per-driver logging while keeping the dependency three sane-ish
	dLogger = logging.GetDriverLogger("common")

	driverConfig := config.Sub("driver")
	// logger.Debug("ttl", driverConfig.GetDuration("cmd_ttl")*time.Second)
	httpClient = &http.Client{Timeout: driverConfig.GetDuration("cmd_ttl") * time.Second}
}

func isRegistered(name string) bool {
	// Check if driver has already been registered
	_, ok := drivers[name]
	return ok
}

func registerDriver(name string, generator func(ip string) BaseModem) {
	// Check if driver has already been registered
	if isRegistered(name) {
		panic(fmt.Sprintf("attempted to register %s twice", name))
	}

	// Register the driver
	drivers[name] = generator
}

func GetModemDriver(model string, host string) (BaseModem, error) {
	if !isRegistered(model) {
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
