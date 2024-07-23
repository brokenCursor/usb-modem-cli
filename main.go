package main

import (
	"fmt"
	
	"github.com/brokenCursor/zte-modem-cli/drivers"
	
	"github.com/alexflint/go-arg"
	"github.com/go-playground/validator/v10"
)

// CLI Definitions //

type SMSArgs struct {
	Action      string `arg:""`
	PhoneNumber string `arg:"-n" help:"Recievers"`
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
	parser := arg.MustParse(&args)

	switch {
	case args.Connection != nil:
		err := validate.Struct(args.Connection)

		if err != nil {
			parser.FailSubcommand("Unknown action", "conn")
		}
		fmt.Printf("changing connection status to %s\n", args.Connection.Action)
	case parser.Subcommand() == nil:
		parser.Fail("Missing or unknown command")
	}
	
	GetModemDriver()

	fmt.Printf("Modem cmd: %s\n", args.Ip)
}
