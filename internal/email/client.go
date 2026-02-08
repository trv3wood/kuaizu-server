package email

import (
	"fmt"
	"os"
	"strconv"

	mail "github.com/wneessen/go-mail"
)

// Client 邮件客户端接口
type Client interface {
	Send(to, subject, htmlBody string) error
}

// SMTPClient SMTP邮件客户端
type SMTPClient struct {
	host     string
	port     int
	user     string
	password string
	fromName string
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	FromName string
}

// NewSMTPClient 创建SMTP客户端
func NewSMTPClient(config SMTPConfig) *SMTPClient {
	return &SMTPClient{
		host:     config.Host,
		port:     config.Port,
		user:     config.User,
		password: config.Password,
		fromName: config.FromName,
	}
}

// NewSMTPClientFromEnv 从环境变量创建SMTP客户端
func NewSMTPClientFromEnv() (*SMTPClient, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil, fmt.Errorf("SMTP_HOST not configured")
	}

	portStr := os.Getenv("SMTP_PORT")
	port := 465 // 默认使用SSL端口
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
		}
	}

	user := os.Getenv("SMTP_USER")
	if user == "" {
		return nil, fmt.Errorf("SMTP_USER not configured")
	}

	password := os.Getenv("SMTP_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD not configured")
	}

	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "快组团队"
	}

	return &SMTPClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		fromName: fromName,
	}, nil
}

// Send 发送邮件
func (c *SMTPClient) Send(to, subject, htmlBody string) error {
	// 构建邮件
	msg := mail.NewMsg()
	if err := msg.FromFormat(c.fromName, c.user); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)

	// 根据端口选择TLS策略
	tlsPolicy := mail.TLSMandatory
	if c.port == 80 || c.port == 25 {
		tlsPolicy = mail.NoTLS
	}

	// 创建go-mail客户端并发送
	client, err := mail.NewClient(c.host,
		mail.WithPort(c.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(c.user),
		mail.WithPassword(c.password),
		mail.WithTLSPolicy(tlsPolicy),
	)
	if err != nil {
		return fmt.Errorf("create mail client: %w", err)
	}

	// 端口465使用SSL
	if c.port == 465 {
		client.SetSSL(true)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}
