package zeropush

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var (
	BASE_URL = "https://api.zeropush.com"
)

type Client struct {
	BaseURL   string
	AuthToken string
}

type DeviceResponse struct {
	*ZeroResponse
	DeviceToken      string
	Active           bool
	MarkedInactiveAt string
	Badge            int
	Channels         []string
}

type NotifyResponse struct {
	*ZeroResponse
	SentCount          int
	InactiveTokens     []string
	UnregisteredTokens []string
}
type BroadcastResponse struct {
	*ZeroResponse
	SentCount int
}

type TokenDetail struct {
	DeviceToken      string
	MarkedInactiveAt string
}

//Responses
type ZeroResponse struct {
	Body    []map[string]interface{}
	Headers map[string][]string
	Error   map[string]string
}
type SuccessResponse struct {
	*ZeroResponse
	Message       string
	AuthTokenType string
}

type TokenResponse struct {
	*ZeroResponse
	TokenDetails []TokenDetail
}

type SubscribeResponse struct {
	*ZeroResponse
	DeviceToken string
	Channels    []string
}

func (zr *ZeroResponse) GetHeader(key string) string {
	if zr.Headers[key] != nil && len(zr.Headers[key]) > 0 {
		return zr.Headers[key][0]
	}
	return ""
}

func NewClient() *Client {
	var auth_token string
	if os.Getenv("ENV") == "production" {
		auth_token = os.Getenv("ZEROPUSH_PROD_TOKEN")
	} else {
		auth_token = os.Getenv("ZEROPUSH_DEV_TOKEN")
	}
	return &Client{BaseURL: BASE_URL, AuthToken: auth_token}
}

func add_authorization(req *http.Request, auth_token string) error {
	if auth_token == "" {
		return errors.New("Auth Token is not set.")
	}
	req.Header.Add("Authorization", `Token token="`+auth_token+`"`)
	return nil
}

func (c *Client) VerifyCredentials() (*SuccessResponse, error) {
	var req *http.Request
	var err error
	if req, err = http.NewRequest("GET", c.BaseURL+"/verify_credentials", nil); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	response, err := send_request(req, false)
	if err != nil {
		return &SuccessResponse{ZeroResponse: response}, err
	}
	return &SuccessResponse{
		ZeroResponse:  response,
		Message:       response.Body[0]["message"].(string),
		AuthTokenType: response.Body[0]["auth_token_type"].(string),
	}, nil
}

func send_request(req *http.Request, expect_array bool) (*ZeroResponse, error) {
	http_client := &http.Client{}
	var res *http.Response
	var err error
	if res, err = http_client.Do(req); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	} else {
		defer res.Body.Close()
		zero_response := &ZeroResponse{}
		zero_response.Headers = res.Header
		decoder := json.NewDecoder(res.Body)
		//200s are success codes
		if res.StatusCode > 299 {
			var e map[string]string
			if err = decoder.Decode(&e); err != nil {
				log.Printf("Error: %s", err)
				return nil, err
			}
			zero_response.Error = e
			err = errors.New(e["error"])
			return zero_response, err
		}
		if expect_array {
			var m []map[string]interface{}
			if err = decoder.Decode(&m); err != nil {
				log.Printf("Error: %s", err)
				return nil, err
			}
			zero_response.Body = m
		} else {
			var m map[string]interface{}
			if err = decoder.Decode(&m); err != nil {
				log.Printf("Error: %s", err)
				return nil, err
			}
			zero_response.Body = make([]map[string]interface{}, 1)
			zero_response.Body[0] = m
		}
		return zero_response, nil
	}

}

func (c *Client) GetInactiveTokens() (*TokenResponse, error) {
	var req *http.Request
	var err error
	if req, err = http.NewRequest("GET", c.BaseURL+"/inactive_tokens", nil); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	response, err := send_request(req, true)
	if err != nil {
		return &TokenResponse{ZeroResponse: response}, err
	}
	var token_details []TokenDetail = make([]TokenDetail, len(response.Body))
	for i, token_detail := range response.Body {
		token_details[i].DeviceToken = token_detail["device_token"].(string)
		token_details[i].MarkedInactiveAt = token_detail["marked_inactive_at"].(string)
	}
	return &TokenResponse{
		ZeroResponse: response,
		TokenDetails: token_details,
	}, nil

}

func (c *Client) GetDevice(device_token string) (*DeviceResponse, error) {
	var req *http.Request
	var err error

	if device_token == "" {
		return nil, errors.New("device token must be set")
	}

	if req, err = http.NewRequest("GET", c.BaseURL+"/devices/"+device_token, nil); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error : %s", err)
		return nil, err
	}
	response, err := send_request(req, false)
	if err != nil {
		return &DeviceResponse{ZeroResponse: response}, err
	}
	var channels []string = make([]string, len(response.Body[0]["channels"].([]interface{})))
	for i, channel := range response.Body[0]["channels"].([]interface{}) {
		channels[i] = channel.(string)
	}
	marked_inactive_at := ""
	if response.Body[0]["marked_inactive_at"] != nil {
		marked_inactive_at = response.Body[0]["marked_inactive_at"].(string)
	}
	return &DeviceResponse{
		ZeroResponse:     response,
		DeviceToken:      response.Body[0]["token"].(string),
		Active:           response.Body[0]["active"].(bool),
		MarkedInactiveAt: marked_inactive_at,
		Badge:            int(response.Body[0]["badge"].(float64)),
		Channels:         channels,
	}, nil
}
func (c *Client) register(device_token string, channel string, register bool) (*SuccessResponse, error) {
	var req *http.Request
	var err error
	if device_token == "" {
		return nil, errors.New("device_token cannot be blank")
	}

	data := url.Values{}
	if device_token != "" {
		data.Set("device_token", device_token)
	}
	if channel != "" {
		data.Add("channel", channel)
	}
	u, _ := url.ParseRequestURI(c.BaseURL)
	u.Path = "/register"
	//are we registering?
	request_type := "POST"
	if !register {
		request_type = "DELETE"
		u.Path = "/unregister"
	}
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	log.Printf("URL: %s", urlStr)
	if req, err = http.NewRequest(request_type, urlStr, nil); err != nil {
		log.Printf("Error creating the request : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error adding the authorization header: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := send_request(req, false)
	if err != nil {
		return &SuccessResponse{ZeroResponse: response}, err
	}
	return &SuccessResponse{
		ZeroResponse: response,
		Message:      response.Body[0]["message"].(string),
	}, nil
}

func (c *Client) SetBadge(device_token string, badge int) (*SuccessResponse, error) {
	var req *http.Request
	var err error
	data := url.Values{}
	if device_token == "" {
		return nil, errors.New("device token cannot be empty")
	}
	if badge < 0 {
		return nil, errors.New("badge should be a positive number")
	}
	data.Set("device_token", device_token)
	data.Add("badge", strconv.Itoa(badge))
	u, _ := url.ParseRequestURI(c.BaseURL)
	u.Path = "/set_badge"
	//are we registering?
	request_type := "POST"
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	log.Printf("URL: %s", urlStr)
	if req, err = http.NewRequest(request_type, urlStr, nil); err != nil {
		log.Printf("Error creating the request : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error adding the authorization header: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := send_request(req, false)
	if err != nil {
		return &SuccessResponse{ZeroResponse: response}, err
	}
	return &SuccessResponse{
		ZeroResponse: response,
		Message:      response.Body[0]["message"].(string),
	}, nil
}
func (c *Client) Notify(alert string, badge string, sound string, info string, expiry string, content_available string, category string, device_tokens ...string) (*NotifyResponse, error) {
	var req *http.Request
	var err error
	if device_tokens == nil || len(device_tokens) == 0 {
		return nil, errors.New("device tokens cannot be empty")
	}

	if alert == "" && info == "" {
		return nil, errors.New("Either alert of info must be set")
	}
	data := url.Values{}
	for _, device_token := range device_tokens {
		if device_token != "" {
			data.Add("device_tokens[]", device_token)
		}
	}
	if alert != "" {
		data.Add("alert", alert)
	}
	if badge != "" {
		data.Add("badge", badge)
	}
	if info != "" {
		data.Add("info", info)
	}
	if expiry != "" {
		data.Add("expiry", expiry)
	}
	if content_available != "" {
		data.Add("content_available", content_available)
	}
	u, _ := url.ParseRequestURI(c.BaseURL)
	u.Path = "/notify"
	request_type := "POST"
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	log.Printf("URL: %s", urlStr)
	if req, err = http.NewRequest(request_type, urlStr, nil); err != nil {
		log.Printf("Error creating the request : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error adding the authorization header: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := send_request(req, false)
	if err != nil {
		return &NotifyResponse{ZeroResponse: response}, err
	}
	var inactive_tokens []string = make([]string, len(response.Body[0]["inactive_tokens"].([]interface{})))
	for i, inactive_token := range response.Body[0]["inactive_tokens"].([]interface{}) {
		inactive_tokens[i] = inactive_token.(string)
	}
	var unregistered_tokens []string = make([]string, len(response.Body[0]["unregistered_tokens"].([]interface{})))
	for i, unregistered_token := range response.Body[0]["unregistered_tokens"].([]interface{}) {
		unregistered_tokens[i] = unregistered_token.(string)
	}
	return &NotifyResponse{
		ZeroResponse:       response,
		InactiveTokens:     inactive_tokens,
		UnregisteredTokens: unregistered_tokens,
		SentCount:          int(response.Body[0]["sent_count"].(float64)),
	}, nil

}

func (c *Client) Broadcast(channel string, alert string, badge string, sound string, info string, expiry string, content_available string, category string) (*BroadcastResponse, error) {
	var req *http.Request
	var err error

	if channel == "" {
		return nil, errors.New("Channel must be set")
	}

	if alert == "" && info == "" {
		return nil, errors.New("Either alert of info must be set")
	}
	data := url.Values{}
	if badge != "" {
		data.Add("badge", badge)
	}
	if info != "" {
		data.Add("info", info)
	}
	if expiry != "" {
		data.Add("expiry", expiry)
	}
	if content_available != "" {
		data.Add("content_available", content_available)
	}
	u, _ := url.ParseRequestURI(c.BaseURL)
	u.Path = "/broadcast/" + channel
	request_type := "POST"
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	log.Printf("URL: %s", urlStr)
	if req, err = http.NewRequest(request_type, urlStr, nil); err != nil {
		log.Printf("Error creating the request : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error adding the authorization header: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := send_request(req, false)
	if err != nil {
		return &BroadcastResponse{ZeroResponse: response}, err
	}
	return &BroadcastResponse{
		ZeroResponse: response,
		SentCount:    int(response.Body[0]["sent_count"].(float64)),
	}, nil
}

func (c *Client) subscribe(device_token string, channel string, sub bool) (*SubscribeResponse, error) {
	var req *http.Request
	var err error
	data := url.Values{}

	if channel == "" {
		err = errors.New("channel is not set")
		return nil, err
	}

	if device_token != "" {
		data.Set("device_token", device_token)
	}
	u, _ := url.ParseRequestURI(c.BaseURL)
	u.Path = "/subscribe/" + channel
	//are we registering?
	request_type := "POST"
	if !sub {
		request_type = "DELETE"
	}
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)
	log.Printf("URL: %s", urlStr)
	if req, err = http.NewRequest(request_type, urlStr, nil); err != nil {
		log.Printf("Error creating the request : %s", err)
		return nil, err
	}
	if err = add_authorization(req, c.AuthToken); err != nil {
		log.Printf("Error adding the authorization header: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := send_request(req, false)
	if err != nil {
		return &SubscribeResponse{ZeroResponse: response}, err
	}
	var channels []string = make([]string, len(response.Body[0]["channels"].([]interface{})))
	for i, channel := range response.Body[0]["channels"].([]interface{}) {
		channels[i] = channel.(string)
	}
	return &SubscribeResponse{
		ZeroResponse: response,
		DeviceToken:  response.Body[0]["device_token"].(string),
		Channels:     channels,
	}, nil

}

func (c *Client) Register(device_token string, channel string) (*SuccessResponse, error) {
	return c.register(device_token, channel, true)
}
func (c *Client) Unregister(device_token string, channel string) (*SuccessResponse, error) {
	return c.register(device_token, channel, false)
}
func (c *Client) Subscribe(device_token string, channel string) (*SubscribeResponse, error) {
	return c.subscribe(device_token, channel, true)
}
func (c *Client) Unsubscribe(device_token string, channel string) (*SubscribeResponse, error) {
	return c.subscribe(device_token, channel, false)
}
