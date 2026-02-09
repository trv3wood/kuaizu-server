package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/internal/response"
)

// GetDashboardStats handles GET /admin/dashboard/stats
func (s *AdminServer) GetDashboardStats(ctx echo.Context) error {
	db := s.repo.DB()
	rctx := ctx.Request().Context()

	var userCount, projectCount, pendingProjectCount, pendingAuthCount, pendingFeedbackCount int64

	if err := db.QueryRowxContext(rctx, "SELECT COUNT(*) FROM `user`").Scan(&userCount); err != nil {
		return response.InternalError(ctx, "failed to count users")
	}
	if err := db.QueryRowxContext(rctx, "SELECT COUNT(*) FROM project WHERE status != 3").Scan(&projectCount); err != nil {
		return response.InternalError(ctx, "failed to count projects")
	}
	if err := db.QueryRowxContext(rctx, "SELECT COUNT(*) FROM project WHERE status = 0").Scan(&pendingProjectCount); err != nil {
		return response.InternalError(ctx, "failed to count pending projects")
	}
	if err := db.QueryRowxContext(rctx, "SELECT COUNT(*) FROM `user` WHERE auth_status = 0 AND auth_img_url IS NOT NULL").Scan(&pendingAuthCount); err != nil {
		return response.InternalError(ctx, "failed to count pending auths")
	}
	if err := db.QueryRowxContext(rctx, "SELECT COUNT(*) FROM feedback WHERE status = 0").Scan(&pendingFeedbackCount); err != nil {
		return response.InternalError(ctx, "failed to count pending feedbacks")
	}

	return response.Success(ctx, map[string]interface{}{
		"user_count":             userCount,
		"project_count":          projectCount,
		"pending_project_count":  pendingProjectCount,
		"pending_auth_count":     pendingAuthCount,
		"pending_feedback_count": pendingFeedbackCount,
	})
}
