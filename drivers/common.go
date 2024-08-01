package drivers

import "fmt"

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
		Up            bool
		Down          bool
		Connecting    bool
		Disconnecting bool
	}
)

var drivers map[string]func(host string) BaseModem = map[string]func(host string) BaseModem{}

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
