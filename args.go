package main

// CLI argument definitions
type (
	SMSActionArgs struct {
		// Send *SMSSendArgs `validate:"-" subcommand:"send" help:"Send SMS"`
		// Read *SMSReadArgs `validate:"-" subcommand:"read" help:"Read SMS"`
	}

	// SMSSendArgs struct {
	// 	PhoneNumber string `validate:"e164" arg:"-n,required" help:"Receiver's phone number"`
	// 	Message     string `validate:"printascii" arg:"-m,required" help:"Message to be sent"`
	// }

	// SMSReadArgs struct {
	// }

	ConnectionArgs struct {
		Action string `arg:"positional,required" help:"up/down/status" validate:"oneof=up down status"`
	}

	BaseArgs struct {
		Connection *ConnectionArgs `validate:"-" arg:"subcommand:conn" help:"Manage cell connection"`
		SMS        *SMSActionArgs  `validate:"-" arg:"subcommand:sms" help:"Manage SMS"`
		Ip         string          `validate:"ipv4" arg:"--ip" help:"Override IP in config file"`
	}
)
