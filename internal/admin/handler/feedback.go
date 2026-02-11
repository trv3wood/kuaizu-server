package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
	adminvo "github.com/trv3wood/kuaizu-server/internal/admin/vo"
	"github.com/trv3wood/kuaizu-server/internal/repository"
	"github.com/trv3wood/kuaizu-server/internal/response"
)

// ListFeedbacks handles GET /admin/feedbacks
func (s *AdminServer) ListFeedbacks(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	size, _ := strconv.Atoi(ctx.QueryParam("size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	params := repository.FeedbackListParams{
		Page: page,
		Size: size,
	}

	if v := ctx.QueryParam("status"); v != "" {
		status, err := strconv.Atoi(v)
		if err != nil {
			return response.BadRequest(ctx, "invalid status")
		}
		params.Status = &status
	}

	feedbacks, total, err := s.repo.Feedback.List(ctx.Request().Context(), params)
	if err != nil {
		return response.InternalError(ctx, "failed to list feedbacks")
	}

	list := make([]adminvo.AdminFeedbackVO, len(feedbacks))
	for i := range feedbacks {
		list[i] = *adminvo.NewAdminFeedbackVO(&feedbacks[i])
	}

	return response.Success(ctx, map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetFeedback handles GET /admin/feedbacks/:id
func (s *AdminServer) GetFeedback(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid feedback id")
	}

	feedback, err := s.repo.Feedback.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return response.InternalError(ctx, "failed to get feedback")
	}
	if feedback == nil {
		return response.NotFound(ctx, "feedback not found")
	}

	return response.Success(ctx, adminvo.NewAdminFeedbackVO(feedback))
}

type replyFeedbackRequest struct {
	AdminReply string `json:"adminReply"`
}

// ReplyFeedback handles PATCH /admin/feedbacks/:id
func (s *AdminServer) ReplyFeedback(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid feedback id")
	}

	var req replyFeedbackRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, "invalid request body")
	}

	if req.AdminReply == "" {
		return response.BadRequest(ctx, "adminReply is required")
	}

	if err := s.repo.Feedback.Reply(ctx.Request().Context(), id, req.AdminReply); err != nil {
		return response.InternalError(ctx, "failed to reply feedback")
	}

	return response.SuccessMessage(ctx, "操作成功")
}
