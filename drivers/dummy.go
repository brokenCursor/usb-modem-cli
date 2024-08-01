package drivers

// DO NOT USE DIRECTLY
type (
	dummy struct {
		ip string
	}
)

func init() {
	registerDriver("dummy", newDummy)
}

func newDummy(ip string) BaseModem {
	// fmt.Println("Dummy driver enabled!")
	return &dummy{ip: ip}
}

func (m *dummy) GetModel() string {
	return "Dummy"
}

func (m *dummy) SetHost(ip string) error {
	m.ip = ip
	return nil
}

func (m *dummy) GetHost() string {
	return m.ip
}
