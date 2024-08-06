package drivers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/brokenCursor/usb-modem-cli/logging"
	"github.com/warthog618/sms/encoding/gsm7"
)

// DO NOT USE DIRECTLY
type (
	zte8810ft struct {
		host     string
		basePath string
	}

	result struct {
		Result string `json:"result"`
	}

	pppConnected struct {
		Connected string `json:"ppp_status"`
	}
)

var (
	httpClient *http.Client
	l          *slog.Logger
)

func init() {
	l = logging.GetDriverLogger("ZTE 8810FT")
	httpClient = &http.Client{Timeout: time.Second * 30}
	registerDriver("ZTE 8810FT", newZTE8810FT)
	l.Debug("driver registered")
}

func newZTE8810FT(host string) BaseModem {
	return &zte8810ft{host: host, basePath: "/goform/goform_set_cmd_process"}
}

func (m *zte8810ft) getBaseURL(path string) *url.URL {
	return &url.URL{Scheme: "http", Host: m.host, Path: path}
}

func (m *zte8810ft) getNewRequest(method string, url *url.URL, headers http.Header) *http.Request {
	headers.Add("Referer", fmt.Sprintf("http://%s/index.html", m.host))

	return &http.Request{
		Proto:  "HTTP/1.1",
		Method: method,
		URL:    url,
		Header: headers,
	}
}

func (m *zte8810ft) SetHost(ip string) error {
	m.host = ip
	return nil
}

func (m *zte8810ft) GetHost() string {
	return m.host
}

func (m *zte8810ft) GetModel() string {
	return "ZTE 8810FT"
}

func (m *zte8810ft) ConnectCell() error {
	// 	GET /goform/goform_set_cmd_process?goformId=CONNECT_NETWORK
	// Prepare everything to make a request
	u := m.getBaseURL("/goform/goform_set_cmd_process")
	query := u.Query()
	query.Add("goformId", "CONNECT_NETWORK")
	u.RawQuery = query.Encode()
	request := m.getNewRequest("GET", u, http.Header{})

	resp, err := httpClient.Do(request)

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
	request := m.getNewRequest("GET", u, http.Header{})

	resp, err := httpClient.Do(request)
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
	// Lines 251-258
	// /goform/goform_get_cmd_process?isTest=False&cmd=ppp_connected,multi_data=1&sms_received_flag_flag=0&sts_received_flag_flag=0&_=<curr_time>

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

	request := m.getNewRequest("GET", u, http.Header{})

	resp, err := httpClient.Do(request)

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
		return &LinkStatus{Up: true}, nil
	case "ppp_connecting":
		return &LinkStatus{Connecting: true}, nil
	case "ppp_disconnecting":
		return &LinkStatus{Disconnecting: true}, nil
	case "ppp_disconnected":
		return &LinkStatus{Down: true}, nil
	default:
		// Unknown link status occurred
		return nil, ErrUnknown
	}
}

func (m *zte8810ft) SendSMS(phone string, message string) error {
	// GET /goform/goform_set_cmd_process?goformId=SEND_SMS
	// Prepare everything to make a request

	// Encode message into GSM-7 as byte array
	encodedMsg, err := gsm7.Encode([]byte(message))
	switch err {

	}
	if err != nil {
		return ActionError{"sms send", err}
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

	request := m.getNewRequest("POST", u, http.Header{
		"Content-Type": {"application/x-www-form-urlencoded", "charset=UTF-8"}})

	// Some Go-level string manipulation
	stringReader := strings.NewReader(query.Encode())
	stringReadCloser := io.NopCloser(stringReader)
	request.Body = stringReadCloser

	resp, err := httpClient.Do(request)

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
