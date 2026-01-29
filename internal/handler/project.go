package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// ListProjects handles GET /projects
func (s *Server) ListProjects(ctx echo.Context, params api.ListProjectsParams) error {
	// Build list params
	listParams := repository.ListParams{
		Page: 1,
		Size: 10,
	}

	if params.Page != nil {
		listParams.Page = *params.Page
	}
	if params.Size != nil {
		listParams.Size = *params.Size
	}
	if listParams.Page < 1 {
		listParams.Page = 1
	}
	if listParams.Size < 1 || listParams.Size > 100 {
		listParams.Size = 10
	}

	listParams.Keyword = params.Keyword
	listParams.SchoolID = params.SchoolId

	if params.Status != nil {
		status := int(*params.Status)
		listParams.Status = &status
	}
	if params.Direction != nil {
		direction := int(*params.Direction)
		listParams.Direction = &direction
	}

	// Query
	projects, total, err := s.repo.Project.List(ctx.Request().Context(), listParams)
	if err != nil {
		return InternalError(ctx, "获取项目列表失败")
	}

	// Convert to VOs
	list := make([]api.ProjectVO, len(projects))
	for i, p := range projects {
		list[i] = *p.ToVO()
	}

	// Build pagination info
	totalPages := int((total + int64(listParams.Size) - 1) / int64(listParams.Size))
	pageInfo := api.PageInfo{
		Page:       &listParams.Page,
		Size:       &listParams.Size,
		Total:      &total,
		TotalPages: &totalPages,
	}

	return Success(ctx, api.ProjectPageResponse{
		List:     &list,
		PageInfo: &pageInfo,
	})
}

// CreateProject handles POST /projects
func (s *Server) CreateProject(ctx echo.Context) error {
	userID := GetUserID(ctx)

	var req api.CreateProjectDTO
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if req.Name == "" {
		return BadRequest(ctx, "项目名称不能为空")
	}
	if req.MemberCount < 1 {
		return BadRequest(ctx, "需求人数必须大于0")
	}

	// Create project
	project := &models.Project{
		CreatorID:       userID,
		Name:            req.Name,
		Description:     &req.Description,
		SchoolID:        req.SchoolId,
		MemberCount:     &req.MemberCount,
		Status:          0, // 待审核
		PromotionStatus: 0, // 无推广
		ViewCount:       0,
	}

	if req.Direction != nil {
		direction := int(*req.Direction)
		project.Direction = &direction
	}

	if err := s.repo.Project.Create(ctx.Request().Context(), project); err != nil {
		return InternalError(ctx, "创建项目失败")
	}

	return Success(ctx, project.ToVO())
}

// GetProject handles GET /projects/{id}
func (s *Server) GetProject(ctx echo.Context, id int) error {
	project, err := s.repo.Project.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取项目详情失败")
	}
	if project == nil {
		return NotFound(ctx, "项目不存在")
	}

	// Increment view count (fire and forget)
	go func() {
		_ = s.repo.Project.IncrementViewCount(ctx.Request().Context(), id)
	}()

	return Success(ctx, project.ToDetailVO())
}

// UpdateProject handles PUT /projects/{id}
func (s *Server) UpdateProject(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)

	// Check ownership
	isOwner, err := s.repo.Project.IsOwner(ctx.Request().Context(), id, userID)
	if err != nil {
		return InternalError(ctx, "检查权限失败")
	}
	if !isOwner {
		return Forbidden(ctx, "只有队长可以修改项目")
	}

	// Get existing project
	project, err := s.repo.Project.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取项目信息失败")
	}
	if project == nil {
		return NotFound(ctx, "项目不存在")
	}

	// Bind request
	var req api.UpdateProjectDTO
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	// Update fields
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}
	if req.Direction != nil {
		direction := int(*req.Direction)
		project.Direction = &direction
	}
	if req.MemberCount != nil {
		project.MemberCount = req.MemberCount
	}

	if err := s.repo.Project.Update(ctx.Request().Context(), project); err != nil {
		return InternalError(ctx, "更新项目失败")
	}

	// Reload project
	project, err = s.repo.Project.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取项目信息失败")
	}

	return Success(ctx, project.ToVO())
}

// DeleteProject handles DELETE /projects/{id}
func (s *Server) DeleteProject(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)

	// Check ownership
	isOwner, err := s.repo.Project.IsOwner(ctx.Request().Context(), id, userID)
	if err != nil {
		return InternalError(ctx, "检查权限失败")
	}
	if !isOwner {
		return Forbidden(ctx, "只有队长可以删除项目")
	}

	if err := s.repo.Project.Delete(ctx.Request().Context(), id); err != nil {
		return InternalError(ctx, "删除项目失败")
	}

	return SuccessMessage(ctx, "项目已删除")
}

// ListProjectApplications handles GET /projects/{id}/applications
func (s *Server) ListProjectApplications(ctx echo.Context, id int, params api.ListProjectApplicationsParams) error {
	return NotImplemented(ctx)
}

// ApplyToProject handles POST /projects/{id}/applications
func (s *Server) ApplyToProject(ctx echo.Context, id int) error {
	return NotImplemented(ctx)
}
