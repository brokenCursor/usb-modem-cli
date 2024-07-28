package main

// CLI argument definitions
type (
	SMSActionArgs struct {
		Send *SMSSendArgs `validate:"-" arg:"subcommand:send" help:"Send SMS"`
		Read *SMSReadArgs `validate:"-" arg:"subcommand:read" help:"Read SMS"`
	}

	SMSSendArgs struct {
		PhoneNumber string `validate:"e164" arg:"-p,--phone,required" help:"Receiver's phone number"`
		Message     string `validate:"printascii" arg:"-m,--msg,required" help:"Message to be sent"`
	}

	SMSReadArgs struct {
	}

	ConnectionArgs struct {
		Action string `arg:"positional,required" help:"up/down/status" validate:"oneof=up down status"`
	}

	BaseArgs struct {
		Connection   *ConnectionArgs `validate:"-" arg:"subcommand:conn" help:"Manage cell connection"`
		SMS          *SMSActionArgs  `validate:"-" arg:"subcommand:sms" help:"Manage SMS"`
		Ip           string          `validate:"omitempty,ipv4" arg:"--ip" help:"Override IP in config file"`
		DisableColor bool            `arg:"-p,--plain" help:"Disable color for better software interaction"`
	}
)
