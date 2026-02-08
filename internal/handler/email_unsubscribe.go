package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// EmailUnsubscribe 处理邮件退订
// GET /api/email/unsubscribe?token=xxx
func (s *Server) EmailUnsubscribe(c echo.Context, params api.EmailUnsubscribeParams) error {
	token := params.Token
	if token == "" {
		return c.HTML(http.StatusBadRequest, unsubscribeErrorHTML("无效的退订链接"))
	}

	// 解析 token 获取 user_id
	userID, err := decodeUnsubscribeToken(token)
	if err != nil {
		return c.HTML(http.StatusBadRequest, unsubscribeErrorHTML("退订链接已失效或无效"))
	}

	// 更新用户退订状态
	if err := s.repo.User.SetEmailOptOut(c.Request().Context(), userID, true); err != nil {
		return c.HTML(http.StatusInternalServerError, unsubscribeErrorHTML("退订失败，请稍后重试"))
	}

	return c.HTML(http.StatusOK, unsubscribeSuccessHTML())
}

// GenerateUnsubscribeToken 生成退订 token
// token 格式: base64(userID:timestamp:signature)
// 有效期: 30天
func GenerateUnsubscribeToken(userID int) string {
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%d:%d", userID, timestamp)

	secretKey := getSecretKey()
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))[:16]

	payload := data + ":" + sig
	return base64.URLEncoding.EncodeToString([]byte(payload))
}

// decodeUnsubscribeToken 解码并验证退订 token
func decodeUnsubscribeToken(token string) (int, error) {
	// 解码 base64
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("invalid token encoding")
	}

	// 解析 payload: userID:timestamp:signature
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token format")
	}

	userIDStr := parts[0]
	timestampStr := parts[1]
	providedSig := parts[2]

	// 验证签名
	data := userIDStr + ":" + timestampStr
	secretKey := getSecretKey()
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	expectedSig := hex.EncodeToString(mac.Sum(nil))[:16]

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return 0, fmt.Errorf("invalid signature")
	}

	// 验证时间戳 (30天有效期)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp")
	}

	tokenTime := time.Unix(timestamp, 0)
	if time.Since(tokenTime) > 30*24*time.Hour {
		return 0, fmt.Errorf("token expired")
	}

	// 解析用户ID
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid user id")
	}

	return userID, nil
}

func getSecretKey() string {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		key = "kuaizu-default-secret-key"
	}
	return key
}

func unsubscribeSuccessHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>退订成功 - 快组</title>
    <style>
        body {
            font-family: 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            margin: 0;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .card {
            background: white;
            border-radius: 16px;
            padding: 40px;
            text-align: center;
            max-width: 400px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            font-size: 24px;
            margin-bottom: 15px;
        }
        p {
            color: #666;
            font-size: 14px;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">✅</div>
        <h1>退订成功</h1>
        <p>您已成功退订邮件推广通知</p>
        <p>如需重新订阅，请在个人中心设置</p>
    </div>
</body>
</html>`
}

func unsubscribeErrorHTML(message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>退订失败 - 快组</title>
    <style>
        body {
            font-family: 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            margin: 0;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .card {
            background: white;
            border-radius: 16px;
            padding: 40px;
            text-align: center;
            max-width: 400px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            font-size: 24px;
            margin-bottom: 15px;
        }
        p {
            color: #666;
            font-size: 14px;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">❌</div>
        <h1>退订失败</h1>
        <p>%s</p>
    </div>
</body>
</html>`, message)
}
