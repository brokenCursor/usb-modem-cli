package dummy

import (
	"log/slog"

	"github.com/brokenCursor/usb-modem-cli/drivers/common"
)

// DO NOT USE DIRECTLY
type (
	dummy struct {
		host string
	}
)

var (
	logger *slog.Logger
)

func init() {
	_, logger = common.RegisterDriver("dummy", newDummy)
}

func newDummy(host string) common.BaseModem {
	logger.With("driver", "dummy").Debug("dummy driver registered")
	return &dummy{host: host}
}

func (m *dummy) GetModel() string {
	return "Dummy"
}

func (m *dummy) SetHost(ip string) error {
	m.host = ip
	return nil
}

func (m *dummy) GetHost() string {
	return m.host
}
