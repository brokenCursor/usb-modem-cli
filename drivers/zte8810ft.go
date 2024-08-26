package drivers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	cfg "github.com/brokenCursor/usb-modem-cli/config"
	"github.com/spf13/viper"
	"github.com/warthog618/sms/encoding/gsm7"
	"github.com/warthog618/sms/encoding/ucs2"
)

// DO NOT USE DIRECTLY
type (
	zte8810ft struct {
		httpClient *http.Client
		logger     *slog.Logger
		config     *viper.Viper
	}

	result struct {
		Result string `json:"result"`
	}

	pppConnected struct {
		Connected string `json:"ppp_status"`
	}

	rawIncomingSMS struct {
		Messages []struct {
			ID           string `json:"id"`
			Source       string `json:"number"`
			Content      string `json:"content"`
			Tag          string `json:"tag"`
			Date         string `json:"date"`
			DraftGroupID string `json:"draft_group_id"`
		} `json:"messages"`
	}
)

func init() {
	RegisterDriver("ZTE 8810FT", newZTE8810FT)
}

func newZTE8810FT(config *viper.Viper, logger *slog.Logger) (BaseModem, error) {
	if config.IsSet("iface") {
		ifaceName := config.GetString("iface")
		logger.With("iface_name", ifaceName).Debug("NIC has been specified")

		// Get this computer's IP on the interface
		ifaceAddr, err := GetInterfaceIPv4Addr(ifaceName)
		if err != nil {
			return nil, cfg.ConfigError{Key: "iface", Value: ifaceName, Err: cfg.ErrInvalidValue}
		}
		logger.With("source_ip", ifaceAddr).Debug("request source ip addr")

		// Create a transport which for the specified interface
		transport, err := GetTransportForIPv4(ifaceAddr)
		if err != nil {
			return nil, err
		}

		return &zte8810ft{logger: logger.With("modem", "ZTE8810FT"), config: config, httpClient: &http.Client{Transport: transport}}, nil
	} else {
		return &zte8810ft{logger: logger.With("modem", "ZTE8810FT"), config: config, httpClient: &http.Client{}}, nil
	}

}

func (m *zte8810ft) getBaseURL(path string) *url.URL {
	return &url.URL{Scheme: "http", Host: m.config.GetString("host"), Path: path}
}

func (m *zte8810ft) getNewRequest(method string, url *url.URL, headers http.Header, body io.Reader) (req *http.Request, err error) {
	headers.Add("Referer", fmt.Sprintf("http://%s/index.html", m.config.GetString("host")))

	req, err = http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header = headers
	return
}

func (m *zte8810ft) GetModel() string {
	return "ZTE 8810FT"
}

func (m *zte8810ft) ConnectCell() error {
	// Prepare everything to make a request
	u := m.getBaseURL("/goform/goform_set_cmd_process")
	query := u.Query()
	query.Add("goformId", "CONNECT_NETWORK")
	u.RawQuery = query.Encode()

	// Create request
	request, err := m.getNewRequest("GET", u, http.Header{}, nil)
	if err != nil {
		return ActionError{Action: "connect", Err: err}
	}
	m.logger.Debug("request", request.URL.String(), nil)

	resp, err := m.httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return ActionError{Action: "connect", Err: err}
	case resp.StatusCode != 200:
		return ActionError{Action: "connect", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return ActionError{Action: "connect", Err: UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return ActionError{Action: "connect", Err: fmt.Errorf("result: %s", result.Result)}
	}

	return nil
}

func (m *zte8810ft) DisconnectCell() error {
	// 	GET /goform/goform_set_cmd_process?goformId=DISCONNECT_NETWORK
	// Prepare everything to make a request
	u := m.getBaseURL("/goform/goform_set_cmd_process")
	query := u.Query()
	query.Add("goformId", "DISCONNECT_NETWORK")
	u.RawQuery = query.Encode()

	// Create request
	request, err := m.getNewRequest("GET", u, http.Header{}, nil)
	if err != nil {
		return ActionError{Action: "disconnect", Err: err}
	}

	m.logger.Debug("request", request.URL.String(), nil)

	resp, err := m.httpClient.Do(request)
	// Process errors
	switch {
	case err != nil:
		return ActionError{Action: "disconnect", Err: err}
	case resp.StatusCode != 200:
		return ActionError{Action: "disconnect", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return ActionError{Action: "disconnect", Err: UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return ActionError{Action: "disconnect", Err: fmt.Errorf("result: %s", result.Result)}
	}

	return nil
}

func (m *zte8810ft) GetCellConnStatus() (*LinkStatus, error) {
	// Build URL
	u := m.getBaseURL("/goform/goform_get_cmd_process")
	query := u.Query()
	query.Add("isTest", "False")
	query.Add("cmd", "ppp_status")
	query.Add("multi_data", "1")
	query.Add("sms_received_flag_flag", "0")
	query.Add("sts_received_flag_flag", "0")
	query.Add("_", strconv.FormatInt((time.Now().UnixMilli)(), 10))
	u.RawQuery = query.Encode()

	request, err := m.getNewRequest("GET", u, http.Header{}, nil)
	if err != nil {
		return nil, ActionError{Action: "status", Err: err}
	}
	m.logger.Debug("request", request.URL.String(), nil)

	resp, err := m.httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return nil, ActionError{Action: "status", Err: err}
	case resp.StatusCode != 200:
		return nil, ActionError{Action: "status", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrUnknown
	}

	result := new(pppConnected)
	if err := json.Unmarshal(body, result); err != nil {
		return nil, ActionError{Action: "status", Err: UnmarshalError{RawData: &body, Err: err}}
	}

	// Process the result
	switch result.Connected {
	case "ppp_connected":
		return &LinkStatus{State: 3}, nil
	case "ppp_connecting":
		return &LinkStatus{State: 2}, nil
	case "ppp_disconnecting":
		return &LinkStatus{State: 1}, nil
	case "ppp_disconnected":
		return &LinkStatus{State: 0}, nil
	default:
		// Unknown link status occurred
		return nil, ErrUnknown
	}
}

func (m *zte8810ft) SendSMS(phone string, message string) error {
	// Encode message into GSM-7 as byte array
	encodedMsg, err := gsm7.Encode([]byte(message))
	if err != nil {
		return ActionError{Action: "sms send", Err: err}
	}

	// Format into proprietary two byte format: 1122 (0074)
	formattedMsg := ""
	for _, hex := range encodedMsg {
		formattedMsg += fmt.Sprintf("00%X", hex)
	}
	u := m.getBaseURL("/goform/goform_set_cmd_process")

	// Build body
	query := u.Query()
	query.Add("goformId", "SEND_SMS")
	query.Add("ID", "-1")
	query.Add("encode_type", "GSM7_default")
	query.Add("Number", phone)
	query.Add("MessageBody", formattedMsg)

	// Build send timestamp
	t := time.Now()
	if _, tz := t.Zone(); tz >= 0 {
		query.Add("sms_time", t.Format("06;01;02;15;04;05;+")+strconv.Itoa(tz/3600))
	} else {
		query.Add("sms_time", t.Format("06;01;02;15;04;05;")+strconv.Itoa(tz/3600))
	}

	// Some Go-level string manipulation
	encoded := query.Encode()
	queryReader := strings.NewReader(encoded)

	// Create request
	request, err := m.getNewRequest("POST", u, http.Header{
		"Content-Type": {"application/x-www-form-urlencoded; charset=UTF-8"}}, queryReader)

	if err != nil {
		return ActionError{Action: "sms send", Err: err}
	}
	m.logger.Debug("url", request.URL.String(), "body", encoded, nil)

	resp, err := m.httpClient.Do(request)
	request.Close = true

	// Process errors
	switch {
	case err != nil:
		return ActionError{Action: "sms send", Err: err}
	case resp.StatusCode != 200:
		return ActionError{Action: "sms send", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return ActionError{Action: "sms send", Err: UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return ActionError{Action: "sms send", Err: fmt.Errorf("result: %s", result.Result)}
	}

	return nil
}

func (m *zte8810ft) ReadAllSMS() ([]SMS, error) {
	// http://10.96.170.1/goform/goform_get_cmd_process?cmd=sms_data_total&page=0&data_per_page=100&mem_store=1&tags=12&order_by=order+by+id+desc&_=1724578532798
	// Build query
	u := m.getBaseURL("/goform/goform_get_cmd_process")
	query := u.Query()
	query.Add("cmd", "sms_data_total")
	query.Add("page", "0")
	query.Add("data_per_page", "100")
	query.Add("mem_store", "1")
	query.Add("tags", "12") // 1 - Received, unread | 2 - Sent | 12 - Received, read + unread | 11 - Drafts (??)
	query.Add("order_by", "order by id desc")
	query.Add("_", strconv.FormatInt((time.Now().UnixMilli)(), 10))
	u.RawQuery = query.Encode()

	request, err := m.getNewRequest("GET", u, http.Header{}, nil)
	if err != nil {
		return nil, ActionError{Action: "sms read", Err: err}
	}
	m.logger.With("request", request.URL.String()).Debug("request")

	resp, err := m.httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return nil, ActionError{Action: "status", Err: err}
	case resp.StatusCode != 200:
		return nil, ActionError{Action: "status", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrUnknown
	}

	rawSMS := new(rawIncomingSMS)

	err = json.Unmarshal(body, &rawSMS)
	if err != nil {
		m.logger.With("body", body).Debug("failed to unmarshal body to struct")
		return nil, ActionError{Action: "sms read", Err: err}
	}

	processedSMS := make([]SMS, len(rawSMS.Messages))
	for i := range rawSMS.Messages {
		processedSMS[i].Sender = rawSMS.Messages[i].Source

		// Extract datetime
		date, err := time.Parse("06,01,02,15,04,05,-07", rawSMS.Messages[i].Date)
		if err != nil {
			m.logger.With("id", rawSMS.Messages[i].ID, "raw_date", rawSMS.Messages[i].Date, "err", err).Debug("failed to parse datetime")
			return nil, ActionError{Action: "sms read", Err: fmt.Errorf("failed to decode message datetime")}
		}
		processedSMS[i].Time = date

		// Extract contents
		rawBytes, err := hex.DecodeString(rawSMS.Messages[i].Content)
		if err != nil {
			m.logger.With("id", rawSMS.Messages[i].ID, "raw_content", rawSMS.Messages[i].Content).Debug("failed to parse content")
			return nil, ActionError{Action: "sms read", Err: fmt.Errorf("failed to parse message content")}
		}

		runes, err := ucs2.Decode(rawBytes)
		if err != nil {
			m.logger.With("id", rawSMS.Messages[i].ID, "raw_content", rawSMS.Messages[i].Content).Debug("failed to decode content")
			return nil, ActionError{Action: "sms read", Err: fmt.Errorf("failed to decode message content")}
		}

		processedSMS[i].Message = string(runes)
	}

	return processedSMS, nil
}
