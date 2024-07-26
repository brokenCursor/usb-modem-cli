package main

import (
	"fmt"
	"os"

	"github.com/brokenCursor/usb-modem-cli/drivers"

	"github.com/alexflint/go-arg"
	"github.com/go-playground/validator/v10"
)

// CLI Definitions //

type SMSArgs struct {
	Action      string `arg:""`
	PhoneNumber string `arg:"-n" help:"Receiver's phone number"`
	Message     string `arg:"-m"`
}

type ConnectionArgs struct {
	Action string `arg:"positional,required" help:"up/down/status" validate:"oneof=up down status"`
}

type BaseArgs struct {
	Connection *ConnectionArgs `validate:"-" arg:"subcommand:conn" help:"Manage cell connection"`
	SMS        *SMSArgs        `validate:"-" arg:"subcommand:sms" help:"Manage SMS"`
	Ip         string          `validate:"ipv4" arg:"--ip" help:"Override IP in config file"`
}

var validate *validator.Validate
var args BaseArgs

func init() {
	// Create a single instance of validator
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func run() error {
	parser := arg.MustParse(&args)
	model := "8810FT"

	modem, err := drivers.GetModemDriver(model, "192.168.0.1")
	if err != nil {
		panic("failed to get drivers")
	}

	switch {
	case args.Connection != nil:
		err := validate.Struct(args.Connection)

		cell, ok := modem.(drivers.ModemCell)
		if !ok {
			return DriverSupportError{Driver: modem, Function: "cell connection"}
		}

		switch args.Connection.Action {
		case "status":
			isConnected, err := cell.GetCellConnStatus()

			if err != nil {
				fmt.Println("Failed to get connection status")
				return
			}
			if isConnected {
				fmt.Println("Status: up")
			} else {
				fmt.Println("Status: down")
			}
		}

		if err != nil {
			parser.FailSubcommand("Unknown action", "conn")
		}
		fmt.Printf("changing connection status to %s\n", args.Connection.Action)
	case parser.Subcommand() == nil:
		parser.Fail("Missing or unknown command")
	}

	fmt.Printf("Modem cmd: %s\n", args.Ip)

	return nil
}
