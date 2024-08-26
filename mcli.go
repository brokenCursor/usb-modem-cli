package main

import (
	"log/slog"
	"os"
	"time"

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

func main() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	logger = logging.GetGeneralLogger()

	if err := run(); err != nil {
		cfmt.Fprintf(os.Stderr, "{{error:}}::red %v\n", err)
	}
}

func run() error {
	parser := arg.MustParse(&args)

	// Get modem configuration and driver
	modemConfig := config.Sub("modem")

	model := modemConfig.GetString("model")

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
		modemConfig.Set("host", args.Host)
	}

	modem, err := drivers.GetModemDriver(model, modemConfig, logging.GetDriverLogger(model))
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
			switch status.State {
			case 0:
				cfmt.Println("Status: {{down}}::red|bold")
			case 1:
				cfmt.Println("Status: {{disconnecting}}::#FA8100|bold")
			case 2:
				cfmt.Println("Status: {{connecting}}::yellow|bold")
			case 3:
				cfmt.Println("Status: {{up}}::green|bold")
			}
		}
	case args.SMS != nil:
		sms, ok := modem.(drivers.ModemSMS)
		if !ok {
			return DriverSupportError{Driver: modem, Function: "SMS"}
		}

		switch {
		case args.SMS.Send != nil:
			err := validate.Struct(args.SMS.Send)
			if err != nil {
				logger.With("err", err.Error()).Debug("sms send validation error")
				parser.FailSubcommand("Unknown values or action", "sms")
			}

			err = sms.SendSMS(args.SMS.Send.PhoneNumber, args.SMS.Send.Message)
			if err != nil {
				return err
			}
		case args.SMS.Read != nil:
			messages, err := sms.ReadAllSMS()
			if err != nil {
				return err
			}

			for i := range messages {
				cfmt.Printf("{{ID:}}::cyan %d\n{{Source:}}::green %s\n{{Time:}}::yellow %s\n{{Text:}}::#FA8100\n%s\n---\n", i, messages[i].Sender, messages[i].Time.Format(time.DateTime), messages[i].Message)
			}
		}
	case parser.Subcommand() == nil:
		parser.Fail("Missing or unknown command")
	}

	return nil
}
