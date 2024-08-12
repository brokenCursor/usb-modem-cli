package zte8810ft

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

	"github.com/brokenCursor/usb-modem-cli/drivers/common"
	"github.com/spf13/viper"
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
	logger     *slog.Logger
	config     *viper.Viper
)

func init() {
	fmt.Println("here")
	config, logger = common.RegisterDriver("ZTE 8810FT", newZTE8810FT)
}

func newZTE8810FT(host string) common.BaseModem {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.GetDuration("cmd_ttl") * time.Second}
	}

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

func (m *zte8810ft) SetHost(host string) error {
	m.host = host
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

	logger.Debug("request", request.URL.String(), nil)

	resp, err := httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return common.ActionError{Action: "connect", Err: err}
	case resp.StatusCode != 200:
		return common.ActionError{Action: "connect", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return common.ActionError{Action: "connect", Err: common.UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return common.ActionError{Action: "connect", Err: fmt.Errorf("result: %s", result.Result)}
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

	logger.Debug("request", request.URL.String(), nil)

	resp, err := httpClient.Do(request)
	// Process errors
	switch {
	case err != nil:
		return common.ActionError{Action: "disconnect", Err: err}
	case resp.StatusCode != 200:
		return common.ActionError{Action: "disconnect", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return common.ActionError{Action: "disconnect", Err: common.UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return common.ActionError{Action: "disconnect", Err: fmt.Errorf("result: %s", result.Result)}
	}

	return nil
}

func (m *zte8810ft) GetCellConnStatus() (*common.LinkStatus, error) {
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

	logger.Debug("request", request.URL.String(), nil)

	resp, err := httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return nil, common.ActionError{Action: "status", Err: err}
	case resp.StatusCode != 200:
		return nil, common.ActionError{Action: "status", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.ErrUnknown
	}

	result := new(pppConnected)
	if err := json.Unmarshal(body, result); err != nil {
		return nil, common.ActionError{Action: "status", Err: common.UnmarshalError{RawData: &body, Err: err}}
	}

	// Process the result
	switch result.Connected {
	case "ppp_connected":
		return &common.LinkStatus{State: 3}, nil
	case "ppp_connecting":
		return &common.LinkStatus{State: 2}, nil
	case "ppp_disconnecting":
		return &common.LinkStatus{State: 1}, nil
	case "ppp_disconnected":
		return &common.LinkStatus{State: 0}, nil
	default:
		// Unknown link status occurred
		return nil, common.ErrUnknown
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
		return common.ActionError{"sms send", err}
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
	encoded := query.Encode()
	stringReader := strings.NewReader(encoded)
	stringReadCloser := io.NopCloser(stringReader)
	request.Body = stringReadCloser

	logger.Debug("url", request.URL.String(), "body", encoded, nil)

	resp, err := httpClient.Do(request)

	// Process errors
	switch {
	case err != nil:
		return common.ActionError{Action: "sms send", Err: err}
	case resp.StatusCode != 200:
		return common.ActionError{Action: "sms send", Err: fmt.Errorf("response status %d", resp.StatusCode)}
	}

	// Read the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.ErrUnknown
	}

	result := new(result)
	if err := json.Unmarshal(body, result); err != nil {
		return common.ActionError{Action: "sms send", Err: common.UnmarshalError{RawData: &body, Err: err}}
	}

	if result.Result != "success" {
		return common.ActionError{Action: "sms send", Err: fmt.Errorf("result: %s", result.Result)}
	}

	return nil
}
