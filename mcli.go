package main

import (
	"log/slog"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/brokenCursor/usb-modem-cli/config"
	"github.com/brokenCursor/usb-modem-cli/drivers"
	"github.com/brokenCursor/usb-modem-cli/logging"
	"github.com/go-playground/validator/v10"
	"github.com/i582/cfmt/cmd/cfmt"
)

var (
	validate *validator.Validate
	args     BaseArgs
	logger   *slog.Logger
)

func init() {
	// Create a single instance of validator
	validate = validator.New(validator.WithRequiredStructEnabled())
	logger = logging.GetGeneralLogger()
}

func main() {
	if err := run(); err != nil {
		cfmt.Fprintf(os.Stderr, "{{error:}}::red %v\n", err)
	}
}

func run() error {
	parser := arg.MustParse(&args)

	// Get modem configuration and driver
	modemConfig := config.Sub("modem")

	model := modemConfig.GetString("model")
	ip := modemConfig.GetString("host")

	if err := validate.Struct(args); err != nil {
		// TODO: add actual error output
		parser.Fail("invalid value for \"--host\" ")
	}

	if args.DisableColor {
		logger.Debug("Colored output disabled")
		cfmt.DisableColors()
	}

	// If IP has been overridden
	if len(args.Host) > 0 {
		logger.Debug("IP has been overridden")
		ip = args.Host
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
		case "up":
			err := cell.ConnectCell()
			if err != nil {
				return err
			}
		case "down":
			err := cell.DisconnectCell()
			if err != nil {
				return err
			}
		case "status":
			status, err := cell.GetCellConnStatus()
			if err != nil {
				return err
			}

			// Process and output status
			switch {
			case status.Up:
				cfmt.Println("Status: {{up}}::green|bold")
			case status.Down:
				cfmt.Println("Status: {{down}}::red|bold")
			case status.Connecting:
				cfmt.Println("Status: {{connecting}}::yellow|bold")
			case status.Disconnecting:
				cfmt.Println("Status: {{disconnecting}}::#FA8100|bold")
			}
		}
	case args.SMS != nil:
		// None of this is implemented :)
		sms, ok := modem.(drivers.ModemSMS)
		if !ok {
			return DriverSupportError{Driver: modem, Function: "SMS"}
		}

		switch {
		case args.SMS.Send != nil:
			err := validate.Struct(args.SMS.Send)
			if err != nil {
				parser.FailSubcommand("Unknown action", "sms")
			}

			err = sms.SendSMS(args.SMS.Send.PhoneNumber, args.SMS.Send.Message)
			if err != nil {
				return err
			}
		}
	case parser.Subcommand() == nil:
		parser.Fail("Missing or unknown command")
	}

	return nil
}
