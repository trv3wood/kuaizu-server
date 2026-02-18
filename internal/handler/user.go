package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// GetCurrentUser handles GET /users/me
func (s *Server) GetCurrentUser(ctx echo.Context) error {
	userID := GetUserID(ctx)

	user, err := s.repo.User.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		return InternalError(ctx, "获取用户信息失败")
	}
	if user == nil {
		return NotFound(ctx, "用户不存在")
	}

	return Success(ctx, user.ToVO())
}

// UpdateCurrentUser handles PUT /users/me
func (s *Server) UpdateCurrentUser(ctx echo.Context) error {
	userID := GetUserID(ctx)

	// Bind request body
	var req api.UpdateUserDTO
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	// Get existing user
	user, err := s.repo.User.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		return InternalError(ctx, "获取用户信息失败")
	}
	if user == nil {
		return NotFound(ctx, "用户不存在")
	}

	// Update fields if provided
	if req.Nickname != nil {
		user.Nickname = req.Nickname
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Email != nil {
		email := string(*req.Email)
		user.Email = &email
	}
	if req.SchoolId != nil {
		user.SchoolID = req.SchoolId
	}
	if req.MajorId != nil {
		user.MajorID = req.MajorId
	}
	if req.Grade != nil {
		user.Grade = req.Grade
	}

	// Save changes
	if err := s.repo.User.Update(ctx.Request().Context(), user); err != nil {
		return InternalError(ctx, "更新用户信息失败")
	}

	// Reload user with joined data
	user, err = s.repo.User.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		return InternalError(ctx, "获取用户信息失败")
	}

	return Success(ctx, user.ToVO())
}

// SubmitCertification handles POST /users/me/certification
func (s *Server) SubmitCertification(ctx echo.Context) error {
	userID := GetUserID(ctx)

	// Bind request body
	var req api.SubmitCertificationJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if req.AuthImgUrl == "" {
		return BadRequest(ctx, "学生证照片不能为空")
	}

	// Get existing user
	user, err := s.repo.User.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		return InternalError(ctx, "获取用户信息失败")
	}
	if user == nil {
		return NotFound(ctx, "用户不存在")
	}

	// Update certification info - 注意新API中AuthStatus: 0-未认证, 1-已认证, 2-认证失败
	// 提交认证图片后设为未认证状态，等待人工审核
	user.AuthImgUrl = &req.AuthImgUrl
	user.AuthStatus = 0 // 未认证，等待审核

	if err := s.repo.User.Update(ctx.Request().Context(), user); err != nil {
		return InternalError(ctx, "提交认证失败")
	}

	return SuccessMessage(ctx, "认证申请已提交，请等待审核")
}

func (s *Server) GetCertificationStatus(ctx echo.Context) error {
	return NotImplemented(ctx)
}
