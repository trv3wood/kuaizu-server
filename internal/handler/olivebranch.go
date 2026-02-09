package handler

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

const dailyFreeQuota = 5

// GetMyReceivedOliveBranches handles GET /users/me/olive-branches
func (s *Server) GetMyReceivedOliveBranches(ctx echo.Context, params api.GetMyReceivedOliveBranchesParams) error {
	userID := GetUserID(ctx)

	listParams := repository.OliveBranchListParams{
		ReceiverID: userID,
		Page:       1,
		Size:       10,
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

	if params.Status != nil {
		status := int(*params.Status)
		listParams.Status = &status
	}

	records, total, err := s.repo.OliveBranch.ListByReceiverID(ctx.Request().Context(), listParams)
	if err != nil {
		return InternalError(ctx, "获取橄榄枝列表失败")
	}

	list := make([]api.OliveBranchVO, len(records))
	for i, ob := range records {
		list[i] = *ob.ToVO()
	}

	totalPages := int((total + int64(listParams.Size) - 1) / int64(listParams.Size))
	pageInfo := api.PageInfo{
		Page:       &listParams.Page,
		Size:       &listParams.Size,
		Total:      &total,
		TotalPages: &totalPages,
	}

	return Success(ctx, api.OliveBranchPageResponse{
		List:     &list,
		PageInfo: &pageInfo,
	})
}

// SendOliveBranch handles POST /olive-branches
func (s *Server) SendOliveBranch(ctx echo.Context) error {
	userID := GetUserID(ctx)

	var req api.SendOliveBranchJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	// Validate receiver exists
	receiver, err := s.repo.User.GetByID(ctx.Request().Context(), req.ReceiverId)
	if err != nil {
		return InternalError(ctx, "查询用户失败")
	}
	if receiver == nil {
		return NotFound(ctx, "接收用户不存在")
	}

	// Cannot send to self
	if req.ReceiverId == userID {
		return BadRequest(ctx, "不能向自己发送橄榄枝")
	}

	// Validate type
	if req.Type != 1 && req.Type != 2 {
		return BadRequest(ctx, "类型无效，必须为1(人才互联)或2(项目邀请)")
	}

	// If project invitation, validate project exists
	var projectName *string
	if req.Type == 2 {
		if req.RelatedProjectId == nil {
			return BadRequest(ctx, "项目邀请必须指定项目ID")
		}
		project, err := s.repo.Project.GetByID(ctx.Request().Context(), *req.RelatedProjectId)
		if err != nil {
			return InternalError(ctx, "查询项目失败")
		}
		if project == nil {
			return NotFound(ctx, "关联项目不存在")
		}
		// Only project creator can send invitation
		if project.CreatorID != userID {
			return Forbidden(ctx, "只有项目队长可以发送邀请")
		}
		projectName = &project.Name
	}

	// Get sender for quota check
	sender, err := s.repo.User.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		return InternalError(ctx, "获取用户信息失败")
	}

	// Determine cost type with daily free quota logic
	today := time.Now().Truncate(24 * time.Hour)
	costType := 0

	// Reset free quota if last active date is not today
	if sender.LastActiveDate == nil || sender.LastActiveDate.Truncate(24*time.Hour).Before(today) {
		sender.FreeBranchUsedToday = 0
		sender.LastActiveDate = &today
	}

	// Check quota: free first, then paid
	if sender.FreeBranchUsedToday < dailyFreeQuota {
		costType = 1 // Free quota
		sender.FreeBranchUsedToday++
	} else if sender.OliveBranchCount > 0 {
		costType = 2 // Paid quota
		sender.OliveBranchCount--
	} else {
		return Error(ctx, 4002, "橄榄枝额度不足，今日免费额度已用完且无付费余额")
	}

	// Update user quota
	if err := s.repo.User.UpdateQuota(ctx.Request().Context(), sender); err != nil {
		return InternalError(ctx, "更新额度失败")
	}

	// Create olive branch record
	hasSms := false
	if req.HasSmsNotify != nil {
		hasSms = *req.HasSmsNotify
	}

	ob := &models.OliveBranch{
		SenderID:         userID,
		ReceiverID:       req.ReceiverId,
		RelatedProjectID: req.RelatedProjectId,
		Type:             req.Type,
		CostType:         costType,
		HasSmsNotify:     hasSms,
		Message:          req.Message,
		Status:           0, // 待处理
	}

	if err := s.repo.OliveBranch.Create(ctx.Request().Context(), ob); err != nil {
		return InternalError(ctx, "发送橄榄枝失败")
	}

	ob.ProjectName = projectName
	ob.Sender = sender
	ob.Receiver = receiver

	return Success(ctx, ob.ToVO())
}

// HandleOliveBranch handles PATCH /olive-branches/{id}
func (s *Server) HandleOliveBranch(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)

	// Get olive branch
	ob, err := s.repo.OliveBranch.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "查询橄榄枝失败")
	}
	if ob == nil {
		return NotFound(ctx, "橄榄枝不存在")
	}

	// Only receiver can handle
	if ob.ReceiverID != userID {
		return Forbidden(ctx, "只有接收者可以处理此邀请")
	}

	// Check current status
	if ob.Status != 0 {
		return BadRequest(ctx, "此邀请已被处理")
	}

	// Bind request
	var req api.HandleOliveBranchJSONBody
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	// Determine new status
	var newStatus int
	switch req.Action {
	case api.ACCEPT:
		newStatus = 1
	case api.REJECT:
		newStatus = 2
	default:
		return BadRequest(ctx, "操作类型无效，必须为ACCEPT或REJECT")
	}

	// Update status
	if err := s.repo.OliveBranch.UpdateStatus(ctx.Request().Context(), id, newStatus); err != nil {
		return InternalError(ctx, "处理邀请失败")
	}

	ob.Status = newStatus
	return Success(ctx, ob.ToVO())
}

// GetMySentOliveBranches handles GET /users/me/sent-olive-branches
func (s *Server) GetMySentOliveBranches(ctx echo.Context, params api.GetMySentOliveBranchesParams) error {
	userID := GetUserID(ctx)

	listParams := repository.OliveBranchListParams{
		SenderID: userID,
		Page:     1,
		Size:     10,
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

	if params.Status != nil {
		status := int(*params.Status)
		listParams.Status = &status
	}

	records, total, err := s.repo.OliveBranch.ListBySenderID(ctx.Request().Context(), listParams)
	if err != nil {
		return InternalError(ctx, "获取橄榄枝列表失败")
	}

	list := make([]api.OliveBranchVO, len(records))
	for i, ob := range records {
		list[i] = *ob.ToVO()
	}

	totalPages := int((total + int64(listParams.Size) - 1) / int64(listParams.Size))
	pageInfo := api.PageInfo{
		Page:       &listParams.Page,
		Size:       &listParams.Size,
		Total:      &total,
		TotalPages: &totalPages,
	}

	return Success(ctx, api.OliveBranchPageResponse{
		List:     &list,
		PageInfo: &pageInfo,
	})
}
