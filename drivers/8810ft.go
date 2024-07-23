package drivers

// DO NOT USE DIRECTLY
type ZTE8810FT struct {
	TargetIP string
	basePath string
}

func init() {
	registerModel("8810FT", newZTE8810FT)
}

func newZTE8810FT(ip string) BaseModem {
	return &ZTE8810FT{TargetIP: ip, basePath: "/goform/goform_set_cmd_process"}
}

func (m *ZTE8810FT) SetTargetIP(ip string) error {
	m.TargetIP = ip
	return nil
}

func (m *ZTE8810FT) GetTargetIP() string {
	return m.TargetIP
}

func (m *ZTE8810FT) GetModel() string {
	return "ZTE 8810FT"
}
