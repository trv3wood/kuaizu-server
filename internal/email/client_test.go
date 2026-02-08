package email

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// TestSMTPClient_SendTestEmail 测试发送邮件到TEST_EMAIL环境变量指定的地址
// 这个测试用于验证SMTP配置是否正确
// 需要设置以下环境变量:
// - SMTP_HOST: SMTP服务器地址
// - SMTP_PORT: SMTP端口 (465或587)
// - SMTP_USER: SMTP用户名
// - SMTP_PASSWORD: SMTP密码
// - SMTP_FROM_NAME: 发件人名称 (可选)
// - TEST_EMAIL: 测试邮件接收地址
func TestSMTPClient_SendTestEmail(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Skip("Warning: .env file not found, using environment variables\n")
	}
	testEmail := os.Getenv("TEST_EMAIL")
	if testEmail == "" {
		t.Skip("TEST_EMAIL environment variable not set, skipping SMTP test")
	}

	client, err := NewSMTPClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create SMTP client from env: %v", err)
	}

	subject := "快组 SMTP 测试邮件"
	htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .success { color: #4CAF50; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>SMTP 测试邮件</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>这是一封来自快组系统的 <span class="success">SMTP 测试邮件</span>。</p>
            <p>如果您收到这封邮件，说明 SMTP 配置正确，邮件服务可以正常使用。</p>
            <p><strong>测试信息：</strong></p>
            <ul>
                <li>发送时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</li>
                <li>SMTP 服务器: ` + client.host + `</li>
                <li>SMTP 端口: ` + fmt.Sprintf("%d", client.port) + `</li>
            </ul>
        </div>
        <div class="footer">
            <p>此邮件由快组系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`

	err = client.Send(testEmail, subject, htmlBody)
	if err != nil {
		t.Logf("\n=== SMTP Configuration ===")
		t.Logf("  Host: %s", client.host)
		t.Logf("  Port: %d", client.port)
		t.Logf("  User: %s", client.user)

		if strings.Contains(err.Error(), "535") {
			t.Logf("\n=== Aliyun DirectMail 535 Authentication Failure ===")
			t.Logf("Common causes and solutions:")
			t.Logf("1. SMTP_USER must be complete email address (e.g., noreply@kuaizu.com)")
			t.Logf("   Current: %s", client.user)
			t.Logf("2. SMTP_PASSWORD must be SMTP password from DirectMail console (not account password)")
			t.Logf("   Go to: Aliyun Console → DirectMail → Sender Addresses → Set SMTP Password")
			t.Logf("3. Sender address must be configured and verified in DirectMail console")
			t.Logf("4. If on Aliyun ECS, use port 80 or 465 (port 25 is blocked)")
			t.Logf("   Current port: %d", client.port)
			t.Logf("5. Verify domain is verified in DirectMail console")
		}

		t.Fatalf("\nFailed to send test email: %v", err)
	}

	t.Logf("✓ Test email sent successfully to %s", testEmail)
}

// TestSMTPClient_SendTestEmail_InvalidRecipient 测试发送到无效邮箱地址
func TestSMTPClient_SendTestEmail_InvalidRecipient(t *testing.T) {
	// 跳过此测试，除非明确需要测试错误处理
	t.Skip("Skipping invalid recipient test")

	client, err := NewSMTPClientFromEnv()
	if err != nil {
		t.Skip("SMTP not configured, skipping test")
	}

	err = client.Send("invalid-email", "Test", "Test body")
	if err == nil {
		t.Error("Expected error when sending to invalid email, got nil")
	}
}

// TestNewSMTPClientFromEnv 测试从环境变量创建客户端
func TestNewSMTPClientFromEnv(t *testing.T) {
	// 保存原始环境变量
	originalHost := os.Getenv("SMTP_HOST")
	originalPort := os.Getenv("SMTP_PORT")
	originalUser := os.Getenv("SMTP_USER")
	originalPassword := os.Getenv("SMTP_PASSWORD")

	// 测试完成后恢复
	defer func() {
		os.Setenv("SMTP_HOST", originalHost)
		os.Setenv("SMTP_PORT", originalPort)
		os.Setenv("SMTP_USER", originalUser)
		os.Setenv("SMTP_PASSWORD", originalPassword)
	}()

	tests := []struct {
		name        string
		setupEnv    func()
		expectError bool
	}{
		{
			name: "missing SMTP_HOST",
			setupEnv: func() {
				os.Unsetenv("SMTP_HOST")
				os.Setenv("SMTP_USER", "test@example.com")
				os.Setenv("SMTP_PASSWORD", "password")
			},
			expectError: true,
		},
		{
			name: "missing SMTP_USER",
			setupEnv: func() {
				os.Setenv("SMTP_HOST", "smtp.example.com")
				os.Unsetenv("SMTP_USER")
				os.Setenv("SMTP_PASSWORD", "password")
			},
			expectError: true,
		},
		{
			name: "missing SMTP_PASSWORD",
			setupEnv: func() {
				os.Setenv("SMTP_HOST", "smtp.example.com")
				os.Setenv("SMTP_USER", "test@example.com")
				os.Unsetenv("SMTP_PASSWORD")
			},
			expectError: true,
		},
		{
			name: "valid configuration",
			setupEnv: func() {
				os.Setenv("SMTP_HOST", "smtp.example.com")
				os.Setenv("SMTP_PORT", "587")
				os.Setenv("SMTP_USER", "test@example.com")
				os.Setenv("SMTP_PASSWORD", "password")
			},
			expectError: false,
		},
		{
			name: "default port when not specified",
			setupEnv: func() {
				os.Setenv("SMTP_HOST", "smtp.example.com")
				os.Unsetenv("SMTP_PORT")
				os.Setenv("SMTP_USER", "test@example.com")
				os.Setenv("SMTP_PASSWORD", "password")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			client, err := NewSMTPClientFromEnv()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected client but got nil")
				}
			}
		})
	}
}
