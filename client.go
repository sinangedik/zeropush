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

type Client struct {
	BaseURL   string
	AuthToken string
}

type ZeroResponse struct {
	Body    []map[string]interface{}
	Headers map[string][]string
	Error   map[string]string
}

func (zr *ZeroResponse) GetHeader(key string) string {
	if zr.Headers[key] != nil && len(zr.Headers[key]) > 0 {
		return zr.Headers[key][0]
	}
	return ""
}

func NewClient() *Client {
	server_address := os.Getenv("BASE_URL")
	var auth_token string
	if os.Getenv("ENV") == "production" {
		auth_token = os.Getenv("ZEROPUSH_PROD_TOKEN")
	} else {
		auth_token = os.Getenv("ZEROPUSH_DEV_TOKEN")
	}
	return &Client{BaseURL: server_address, AuthToken: auth_token}
}

func add_authorization(req *http.Request, auth_token string) error {
	if auth_token == "" {
		return errors.New("Auth Token is not set.")
	}
	req.Header.Add("Authorization", `Token token="`+auth_token+`"`)
	return nil
}

func (c *Client) VerifyCredentials() (*ZeroResponse, error) {
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
	return send_request(req, false)
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

func (c *Client) GetInactiveTokens() (*ZeroResponse, error) {
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
	return send_request(req, true)
}

func (c *Client) GetDevice(device_token string) (*ZeroResponse, error) {
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
	return send_request(req, false)
}
func (c *Client) register(device_token string, channel string, register bool) (*ZeroResponse, error) {
	var req *http.Request
	var err error
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
	return send_request(req, false)
}

func (c *Client) SetBadge(device_token string, badge int) (*ZeroResponse, error) {
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
	return send_request(req, false)
}
func (c *Client) Notify(alert string, badge string, sound string, info string, expiry string, content_available string, category string, device_tokens ...string) (*ZeroResponse, error) {
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
	return send_request(req, false)
}

func (c *Client) Broadcast(channel string, alert string, badge string, sound string, info string, expiry string, content_available string, category string) (*ZeroResponse, error) {
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
	return send_request(req, false)
}

func (c *Client) subscribe(device_token string, channel string, sub bool) (*ZeroResponse, error) {
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
	return send_request(req, false)
}

func (c *Client) Register(device_token string, channel string) (*ZeroResponse, error) {
	return c.register(device_token, channel, true)
}
func (c *Client) Unregister(device_token string, channel string) (*ZeroResponse, error) {
	return c.register(device_token, channel, false)
}
func (c *Client) Subscribe(device_token string, channel string) (*ZeroResponse, error) {
	return c.subscribe(device_token, channel, true)
}
func (c *Client) Unsubscribe(device_token string, channel string) (*ZeroResponse, error) {
	return c.subscribe(device_token, channel, false)
}
