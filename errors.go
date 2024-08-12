package main

import (
	"fmt"

	"github.com/brokenCursor/usb-modem-cli/drivers/common"
)

type DriverSupportError struct {
	Driver   common.BaseModem
	Function string
}

func (e DriverSupportError) Error() string {
	return fmt.Sprintf("driver \"%s\" does not support %s", e.Driver.GetModel(), e.Function)
}
