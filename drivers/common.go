package drivers

import "fmt"

type BaseModem interface {
	SetTargetIP(ip string) error
	GetTargetIP() string
	GetModel() string
}

type ModemCell interface {
	BaseModem

	GetCellConnStatus() (bool, error)
	ConnectCell() error
	DisconnectCell() error
}

type SMS struct {
	Sender  string
	Message string
}

type ModemSMS interface {
	BaseModem

	SendSMS(phone string, message string) error
	GetAllSMS() ([]SMS, error)
}

var models map[string]func(ip string) BaseModem = map[string]func(ip string) BaseModem{}

func isRegistered(name string) bool {
	// Check if model has already been registered
	_, ok := models[name]
	return ok
}

func registerModel(name string, generator func(ip string) BaseModem) {
	// Check if model has already been registered
	if isRegistered(name) {
		panic(fmt.Sprintf("attempted to register %s twice", name))
	}

	// Register the model
	models[name] = generator
}

func GetModemDriver(model string, ip string) (BaseModem, error) {
	if !isRegistered(model) {
		return nil, ErrUnknownModel
	}

	return models[model](ip), nil
}

func GetAvailableDrivers() *[]string {
	return nil
}
