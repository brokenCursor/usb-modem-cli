package drivers


import (
	"net/http"
)

// DO NOT USE DIRECTLY
type zte8810ft struct {
	TargetIP string
	basePath string
}

func init() {
	registerModel("8810FT", newZTE8810FT)
}

func newZTE8810FT(ip string) BaseModem {
	return &zte8810ft{TargetIP: ip, basePath: "/goform/goform_set_cmd_process"}
}

func (m *zte8810ft) SetTargetIP(ip string) error {
	m.TargetIP = ip
	return nil
}

func (m *zte8810ft) GetTargetIP() string {
	return m.TargetIP
}

func (m *zte8810ft) GetModel() string {
	return "ZTE 8810FT"
}

func (m *zte8810ft) ConnectCell() error {
	// 	GET /goform/goform_set_cmd_process?goformId=CONNECT_NETWORK
	return nil
}

func (m *zte8810ft) DisconnectCell() error {
	// 	GET /goform/goform_set_cmd_process?goformId=DISCONNECT_NETWORK
	return nil
}

func (m *zte8810ft) GetCellConnStatus() (bool, error) {
	// Lines 251-258
	resp, err := http.Get()
	// /goform/goform_get_cmd_process?isTest=False&cmd=ppp_connected,multi_data=1&sms_received_flag_flag=0&sts_received_flag_flag=0&_=<curr_time>
	return false, nil
}
