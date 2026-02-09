package wechat

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// PayConfig holds WeChat Pay configuration
type PayConfig struct {
	MchID                string // 商户号
	MchSerialNo          string // 商户证书序列号
	MchAPIKey            string // 商户APIv3密钥
	PrivateKey           string // 商户API私钥(PEM格式)
	AppID                string // 小程序AppID
	NotifyURL            string // 支付回调地址
	WechatPayPublicKey   string // 微信支付公钥(PEM格式)
	WechatPayPublicKeyID string // 微信支付公钥ID
}

// PayClient is the WeChat Pay API client
type PayClient struct {
	config          *PayConfig
	client          *core.Client
	jsapiSvc        *jsapi.JsapiApiService
	notifyHandler   *notify.Handler
	wechatPayPubKey *rsa.PublicKey
}

// PaymentParams 小程序支付参数
type PaymentParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// DefaultPayConfig returns configuration from environment variables
func DefaultPayConfig() (*PayConfig, error) {
	mchID := os.Getenv("WECHAT_MCH_ID")
	mchSerialNo := os.Getenv("WECHAT_MCH_SERIAL_NO")
	mchAPIKey := os.Getenv("WECHAT_MCH_API_KEY")
	privateKeyInput := os.Getenv("WECHAT_MCH_PRIVATE_KEY")
	appID := os.Getenv("WECHAT_APPID")
	notifyURL := os.Getenv("WECHAT_NOTIFY_URL")
	wechatPayPublicKeyInput := os.Getenv("WECHAT_PAY_PUBLIC_KEY")
	wechatPayPublicKeyID := os.Getenv("WECHAT_PAY_PUBLIC_KEY_ID")

	if mchID == "" || mchAPIKey == "" || appID == "" || mchSerialNo == "" {
		return nil, fmt.Errorf("missing required WeChat Pay configuration (WECHAT_MCH_ID, WECHAT_MCH_SERIAL_NO, WECHAT_MCH_API_KEY, WECHAT_APPID)")
	}

	if wechatPayPublicKeyInput == "" || wechatPayPublicKeyID == "" {
		return nil, fmt.Errorf("missing required WeChat Pay public key configuration (WECHAT_PAY_PUBLIC_KEY, WECHAT_PAY_PUBLIC_KEY_ID)")
	}

	// Load private key - can be PEM content, file path, or base64 encoded PEM
	privateKeyPEM, err := loadPEM(privateKeyInput, "private key")
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	// Load WeChat Pay public key - can be PEM content, file path, or base64 encoded PEM
	wechatPayPublicKeyPEM, err := loadPEM(wechatPayPublicKeyInput, "wechat pay public key")
	if err != nil {
		return nil, fmt.Errorf("load wechat pay public key: %w", err)
	}

	return &PayConfig{
		MchID:                mchID,
		MchSerialNo:          mchSerialNo,
		MchAPIKey:            mchAPIKey,
		PrivateKey:           privateKeyPEM,
		AppID:                appID,
		NotifyURL:            notifyURL,
		WechatPayPublicKey:   wechatPayPublicKeyPEM,
		WechatPayPublicKeyID: wechatPayPublicKeyID,
	}, nil
}

// loadPEM loads PEM content from various sources:
// 1. File path - if the path exists as a file
// 2. PEM content - if the string starts with "-----BEGIN"
// 3. Base64 encoded PEM - otherwise try to decode as base64
func loadPEM(input string, name string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("%s input is empty", name)
	}

	// Check if it's a file path
	if _, err := os.Stat(input); err == nil {
		data, err := os.ReadFile(input)
		if err != nil {
			return "", fmt.Errorf("read %s file: %w", name, err)
		}
		return string(data), nil
	}

	// Check if it's already PEM format
	if strings.HasPrefix(strings.TrimSpace(input), "-----BEGIN") {
		return input, nil
	}

	// Try to decode as base64
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 %s: %w", name, err)
	}

	decodedStr := string(decoded)
	if !strings.HasPrefix(strings.TrimSpace(decodedStr), "-----BEGIN") {
		return "", fmt.Errorf("decoded base64 content for %s is not a valid PEM format", name)
	}

	return decodedStr, nil
}

// NewPayClient creates a new WeChat Pay client using official SDK with public key mode
func NewPayClient(config *PayConfig) (*PayClient, error) {
	ctx := context.Background()

	// 加载商户私钥
	privateKey, err := utils.LoadPrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	// 加载微信支付公钥
	wechatPayPublicKey, err := utils.LoadPublicKey(config.WechatPayPublicKey)
	if err != nil {
		return nil, fmt.Errorf("load wechat pay public key: %w", err)
	}

	// 使用公钥模式初始化 client
	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(
			config.MchID,
			config.MchSerialNo,
			privateKey,
			config.WechatPayPublicKeyID,
			wechatPayPublicKey,
		),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create wechat pay client: %w", err)
	}

	// 初始化 jsapi 服务
	jsapiSvc := &jsapi.JsapiApiService{Client: client}

	// 使用公钥验证器创建回调处理器
	pubKeyVerifier := verifiers.NewSHA256WithRSAPubkeyVerifier(config.WechatPayPublicKeyID, *wechatPayPublicKey)
	notifyHandler, err := notify.NewRSANotifyHandler(config.MchAPIKey, pubKeyVerifier)
	if err != nil {
		return nil, fmt.Errorf("create notify handler: %w", err)
	}

	return &PayClient{
		config:          config,
		client:          client,
		jsapiSvc:        jsapiSvc,
		notifyHandler:   notifyHandler,
		wechatPayPubKey: wechatPayPublicKey,
	}, nil
}

// CreatePrepayOrderWithPayment creates a prepay order and returns payment params directly
func (c *PayClient) CreatePrepayOrderWithPayment(ctx context.Context, outTradeNo, description, openID string, amountCents int) (*PaymentParams, error) {
	// 使用 PrepayWithRequestPayment 一次性获取prepay_id和调起支付所需参数
	resp, _, err := c.jsapiSvc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(c.config.AppID),
		Mchid:       core.String(c.config.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(c.config.NotifyURL),
		Amount: &jsapi.Amount{
			Total:    core.Int64(int64(amountCents)),
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(openID),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("prepay: %w", err)
	}

	return &PaymentParams{
		TimeStamp: *resp.TimeStamp,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
	}, nil
}

// ParseNotification parses and verifies the payment notification
func (c *PayClient) ParseNotification(ctx context.Context, request *http.Request) (*payments.Transaction, error) {
	transaction := new(payments.Transaction)
	_, err := c.notifyHandler.ParseNotifyRequest(ctx, request, transaction)
	if err != nil {
		return nil, fmt.Errorf("parse notify: %w", err)
	}

	return transaction, nil
}

// GenerateOutTradeNo generates a unique order number
// Format: KZ{timestamp}_{orderID} to ensure minimum 6 bytes and uniqueness
func GenerateOutTradeNo(orderID int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("KZ%d_%d", timestamp, orderID)
}

// ParseOrderIDFromOutTradeNo parses order ID from out_trade_no
func ParseOrderIDFromOutTradeNo(outTradeNo string) (int, error) {
	// Handle new format: KZ{timestamp}_{orderID}
	if strings.Contains(outTradeNo, "_") {
		parts := strings.Split(outTradeNo, "_")
		if len(parts) != 2 || !strings.HasPrefix(parts[0], "KZ") {
			return 0, fmt.Errorf("invalid out_trade_no format: %s", outTradeNo)
		}
		var orderID int
		_, err := fmt.Sscanf(parts[1], "%d", &orderID)
		if err != nil {
			return 0, fmt.Errorf("invalid out_trade_no format: %s", outTradeNo)
		}
		return orderID, nil
	}

	// Fallback: old format KZ{orderID}
	var orderID int
	_, err := fmt.Sscanf(outTradeNo, "KZ%d", &orderID)
	if err != nil {
		return 0, fmt.Errorf("invalid out_trade_no format: %s", outTradeNo)
	}
	return orderID, nil
}
