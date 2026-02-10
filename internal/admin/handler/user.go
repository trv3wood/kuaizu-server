package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
	adminvo "github.com/trv3wood/kuaizu-server/internal/admin/vo"
	"github.com/trv3wood/kuaizu-server/internal/repository"
	"github.com/trv3wood/kuaizu-server/internal/response"
)

// ListUsers handles GET /admin/users
func (s *AdminServer) ListUsers(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	size, _ := strconv.Atoi(ctx.QueryParam("size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	params := repository.UserListParams{
		Page: page,
		Size: size,
	}

	if v := ctx.QueryParam("authStatus"); v != "" {
		status, err := strconv.Atoi(v)
		if err != nil {
			return response.BadRequest(ctx, "invalid authStatus")
		}
		params.AuthStatus = &status
	}

	if v := ctx.QueryParam("schoolId"); v != "" {
		schoolID, err := strconv.Atoi(v)
		if err != nil {
			return response.BadRequest(ctx, "invalid schoolId")
		}
		params.SchoolID = &schoolID
	}

	if v := ctx.QueryParam("keyword"); v != "" {
		params.Keyword = &v
	}

	users, total, err := s.repo.User.ListUsers(ctx.Request().Context(), params)
	if err != nil {
		return response.InternalError(ctx, "failed to list users")
	}

	list := make([]adminvo.AdminUserVO, len(users))
	for i := range users {
		list[i] = *adminvo.NewAdminUserVO(&users[i])
	}

	return response.Success(ctx, map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetUser handles GET /admin/users/:id
func (s *AdminServer) GetUser(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid user id")
	}

	user, err := s.repo.User.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return response.InternalError(ctx, "failed to get user")
	}
	if user == nil {
		return response.NotFound(ctx, "user not found")
	}

	return response.Success(ctx, adminvo.NewAdminUserVO(user))
}

type reviewAuthRequest struct {
	AuthStatus int `json:"authStatus"`
}

// ReviewUserAuth handles PATCH /admin/users/:id/auth
func (s *AdminServer) ReviewUserAuth(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid user id")
	}

	var req reviewAuthRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, "invalid request body")
	}

	if req.AuthStatus != 1 && req.AuthStatus != 2 {
		return response.BadRequest(ctx, "invalid authStatus, must be 1 (approve) or 2 (reject)")
	}

	if err := s.repo.User.UpdateAuthStatus(ctx.Request().Context(), id, req.AuthStatus); err != nil {
		return response.InternalError(ctx, "failed to update auth status")
	}

	return response.SuccessMessage(ctx, "操作成功")
}
