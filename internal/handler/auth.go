package handler

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/auth"
	"github.com/trv3wood/kuaizu-server/internal/wechat"
)

// LoginWithWechat handles POST /auth/login/wechat
// This endpoint handles both login and registration:
// - If user exists: returns token
// - If user doesn't exist: creates user and returns token
func (s *Server) LoginWithWechat(ctx echo.Context) error {
	// Bind request body
	var req api.LoginWithWechatJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if req.Code == "" {
		return BadRequest(ctx, "微信登录code不能为空")
	}

	// Call WeChat API to get openid
	wxClient := wechat.NewClient()
	wxResp, err := wxClient.Code2Session(req.Code)
	if err != nil {
		log.Printf("WeChat code2session error: %v", err)
		return Error(ctx, 4001, "微信登录失败: "+err.Error())
	}

	// Check if user exists
	user, err := s.repo.User.GetByOpenID(ctx.Request().Context(), wxResp.OpenID)
	if err != nil {
		log.Printf("Get user by openid error: %v", err)
		return InternalError(ctx, "查询用户失败")
	}

	isNewUser := false

	// If user doesn't exist, create new user
	if user == nil {
		isNewUser = true
		user, err = s.repo.User.Create(ctx.Request().Context(), wxResp.OpenID)
		if err != nil {
			log.Printf("Create user error: %v", err)
			return InternalError(ctx, "创建用户失败")
		}
		log.Printf("New user created: id=%d, openid=%s", user.ID, wxResp.OpenID)
	}

	// Generate JWT token
	jwtConfig := auth.DefaultConfig()
	token, expiresIn, err := auth.GenerateToken(jwtConfig, user.ID, wxResp.OpenID)
	if err != nil {
		log.Printf("Generate token error: %v", err)
		return InternalError(ctx, "生成Token失败")
	}

	// Build response
	response := api.LoginResponse{
		Token:     &token,
		ExpiresIn: &expiresIn,
		IsNewUser: &isNewUser,
		User:      user.ToVO(),
	}

	return Success(ctx, response)
}
