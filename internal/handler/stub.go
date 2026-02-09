package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// ========== Commons Module (Not Implemented) ==========

// UploadFile handles POST /commons/uploads
func (s *Server) UploadFile(ctx echo.Context) error {
	return NotImplemented(ctx)
}

// Helper function to check if JSON Body parsing failed
func (s *Server) reviewApplicationHelper(ctx echo.Context, id int) error {
	// 1. Get current user
	userID := GetUserID(ctx)

	// 2. Parse request body
	var req api.ReviewApplicationJSONBody
	if err := ctx.Bind(&req); err != nil {
		return InvalidParams(ctx, err)
	}

	// 3. Validate status
	if req.Status != api.ApplicationStatusN1 && req.Status != api.ApplicationStatusN2 {
		return InvalidParams(ctx, fmt.Errorf("invalid status"))
	}

	// 4. Get application to check existence and project owner
	app, err := s.repo.Application.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "failed to get application")
	}
	if app == nil {
		return NotFound(ctx, "Application not found")
	}

	// 5. Check if current user is the project creator
	isOwner, err := s.repo.Project.IsOwner(ctx.Request().Context(), app.ProjectID, userID)
	if err != nil {
		return InternalError(ctx, "failed to check permission")
	}
	if !isOwner {
		return Forbidden(ctx, "Only project creator can review applications")
	}

	// 6. Update status
	err = s.repo.Application.UpdateStatus(ctx.Request().Context(), id, int(req.Status), req.ReplyMsg)
	if err != nil {
		return InternalError(ctx, "failed to update application status")
	}

	return Success(ctx, nil)
}
