package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/brokenCursor/usb-modem-cli/drivers"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var (
	validate *validator.Validate
	args     BaseArgs
	config   *viper.Viper
)

func init() {
	// Create a single instance of validator
	validate = validator.New(validator.WithRequiredStructEnabled())

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
	config.AddConfigPath(dir + sep + "modem-cli" + sep + "config.yaml")

	config.SetDefault("modem.model", "dummy")
	config.SetDefault("modem.ip", "127.0.0.1")

}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func run() error {
	parser := arg.MustParse(&args)

	// Get modem configuration and driver
	modemConfig := config.Sub("modem")

	model := modemConfig.GetString("model")
	ip := modemConfig.GetString("ip")

	// if args.Ip != "" {
	// 	ip = args.Ip
	// }

	if err := validate.Struct(args); err != nil {
		// TODO: add actual error output
		fmt.Printf("%s\n%v\n", err.Error(), ip)
		parser.Fail("invalid value for \"--ip\" ")
	}

	modem, err := drivers.GetModemDriver(model, ip)
	if err != nil {
		return err
	}

	switch {
	case args.Connection != nil:
		err := validate.Struct(args.Connection)
		if err != nil {
			parser.FailSubcommand("Unknown action", "conn")
		}

		cell, ok := modem.(drivers.ModemCell)
		if !ok {
			return DriverSupportError{Driver: modem, Function: "cell connection"}
		}

		switch args.Connection.Action {
		case "status":
			isConnected, err := cell.GetCellConnStatus()

			if err != nil {
				return err
			}

			if isConnected {
				fmt.Println("Status: up")
			} else {
				fmt.Println("Status: down")
			}
			return nil
		}
	// case args.SMS != nil:
	// 	// None of this is implemented :)
	// 	sms, ok := modem.(drivers.ModemSMS)
	// 	if !ok {
	// 		return DriverSupportError{Driver: modem, Function: "SMS"}
	// 	}

	// 	switch {
	// 	case args.SMS.Read != nil:
	// 		err := validate.Struct(args.SMS.Send)
	// 		if err != nil {
	// 			parser.FailSubcommand("Unknown action", "sms")
	// 		}
	// 		sms.SendSMS(args.SMS.Send.PhoneNumber, args.SMS.Send.PhoneNumber)
	// 	}
	case parser.Subcommand() == nil:
		parser.Fail("Missing or unknown command")
	}

	fmt.Printf("Modem cmd: %s\n", args.Ip)

	return nil
}
