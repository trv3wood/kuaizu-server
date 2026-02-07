package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Code2SessionResponse is the response from WeChat code2session API
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// Client is a WeChat Mini Program API client
type Client struct {
	appID      string
	appSecret  string
	httpClient *http.Client
}

// NewClient creates a new WeChat client from environment variables
func NewClient() *Client {
	return &Client{
		appID:     os.Getenv("WECHAT_APPID"),
		appSecret: os.Getenv("WECHAT_SECRET"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewClientWithConfig creates a new WeChat client with explicit config
func NewClientWithConfig(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Code2Session exchanges the login code for openid and session_key
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (c *Client) Code2Session(code string) (*Code2SessionResponse, error) {
	if c.appID == "" || c.appSecret == "" {
		return nil, fmt.Errorf("WECHAT_APPID or WECHAT_SECRET not configured")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.appID, c.appSecret, code,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request wechat api: %w", err)
	}
	defer resp.Body.Close()

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Check for WeChat API errors
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	if result.OpenID == "" {
		return nil, fmt.Errorf("wechat api returned empty openid")
	}

	return &result, nil
}
