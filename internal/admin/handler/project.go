package handler

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	adminvo "github.com/trv3wood/kuaizu-server/internal/admin/vo"
	"github.com/trv3wood/kuaizu-server/internal/repository"
	"github.com/trv3wood/kuaizu-server/internal/response"
)

// ListProjects handles GET /admin/projects
func (s *AdminServer) ListProjects(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	size, _ := strconv.Atoi(ctx.QueryParam("size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	params := repository.ListParams{
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

	if v := ctx.QueryParam("keyword"); v != "" {
		params.Keyword = &v
	}

	projects, total, err := s.repo.Project.List(ctx.Request().Context(), params)
	if err != nil {
		return response.InternalError(ctx, "failed to list projects")
	}

	list := make([]adminvo.AdminProjectVO, len(projects))
	for i := range projects {
		list[i] = *adminvo.NewAdminProjectVO(&projects[i])
	}

	return response.Success(ctx, map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetProject handles GET /admin/projects/:id
func (s *AdminServer) GetProject(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid project id")
	}

	project, err := s.repo.Project.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return response.InternalError(ctx, "failed to get project")
	}
	if project == nil {
		return response.NotFound(ctx, "project not found")
	}

	return response.Success(ctx, adminvo.NewAdminProjectVO(project))
}

type reviewProjectRequest struct {
	Status int `json:"status"`
}

// ReviewProject handles PATCH /admin/projects/:id
func (s *AdminServer) ReviewProject(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return response.BadRequest(ctx, "invalid project id")
	}

	var req reviewProjectRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, "invalid request body")
	}

	if req.Status != 1 && req.Status != 2 {
		return response.BadRequest(ctx, fmt.Sprintf("invalid status %d, must be 1 (approve) or 2 (reject)", req.Status))
	}

	if err := s.repo.Project.UpdateStatus(ctx.Request().Context(), id, req.Status); err != nil {
		return response.InternalError(ctx, "failed to update project status")
	}

	return response.SuccessMessage(ctx, "操作成功")
}
