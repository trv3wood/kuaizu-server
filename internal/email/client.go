package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
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
	from := fmt.Sprintf("%s <%s>", c.fromName, c.user)

	// 构建邮件内容
	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	// 使用TLS连接 (端口465)
	if c.port == 465 {
		return c.sendWithTLS(addr, to, message)
	}

	// 使用Plain SMTP (端口80或25，用于Aliyun DirectMail)
	if c.port == 80 || c.port == 25 {
		return c.sendWithPlainSMTP(addr, to, message)
	}

	// 使用STARTTLS (端口587)
	return c.sendWithSTARTTLS(addr, to, message)
}

// sendWithTLS 使用TLS发送邮件 (端口465)
func (c *SMTPClient) sendWithTLS(addr, to, message string) error {
	tlsConfig := &tls.Config{
		ServerName: c.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", c.user, c.password, c.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	// 发送邮件
	if err := client.Mail(c.user); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return client.Quit()
}

// sendWithPlainSMTP 使用Plain SMTP发送邮件 (端口80或25，用于Aliyun DirectMail)
func (c *SMTPClient) sendWithPlainSMTP(addr, to, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	// 对于端口80，Aliyun DirectMail要求使用LOGIN认证方式
	// smtp.PlainAuth 默认拒绝在非加密连接上发送凭据
	// 我们需要使用自定义的认证方式或者跳过这个检查
	if c.port == 80 {
		// 使用PLAIN认证，但允许非加密连接
		if err := client.Auth(&plainAuthAllowUnencrypted{
			identity: "",
			username: c.user,
			password: c.password,
			host:     c.host,
		}); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	} else {
		// 端口25使用标准PlainAuth
		auth := smtp.PlainAuth("", c.user, c.password, c.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	// 发送邮件
	if err := client.Mail(c.user); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return client.Quit()
}

// plainAuthAllowUnencrypted 允许在非加密连接上使用PLAIN认证
// 用于Aliyun DirectMail的端口80
type plainAuthAllowUnencrypted struct {
	identity string
	username string
	password string
	host     string
}

func (a *plainAuthAllowUnencrypted) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// 允许非TLS连接
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuthAllowUnencrypted) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}

// sendWithSTARTTLS 使用STARTTLS发送邮件 (端口587)
func (c *SMTPClient) sendWithSTARTTLS(addr, to, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	// STARTTLS
	tlsConfig := &tls.Config{
		ServerName: c.host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	// 认证
	auth := smtp.PlainAuth("", c.user, c.password, c.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	// 发送邮件
	if err := client.Mail(c.user); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return client.Quit()
}
