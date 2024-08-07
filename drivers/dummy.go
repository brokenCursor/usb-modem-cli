package drivers

// DO NOT USE DIRECTLY
type (
	dummy struct {
		host string
	}
)

func init() {
	registerDriver("dummy", newDummy)
	dLogger.With("driver", "dummy").Debug("dummy driver registered")
}

func newDummy(host string) BaseModem {
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
