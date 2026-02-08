package handler

import (
	"context"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/internal/email"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// TriggerEmailPromotionRequest 触发邮件推广请求
type TriggerEmailPromotionRequest struct {
	OrderID   int `json:"orderId"`
	ProjectID int `json:"projectId"`
}

// TriggerEmailPromotion 触发邮件推广
// POST /api/email/promotion/trigger
// 用户购买邮件推广商品后，手动选择要推广的项目
func (s *Server) TriggerEmailPromotion(ctx echo.Context) error {
	userID := GetUserID(ctx)

	var req TriggerEmailPromotionRequest
	if err := ctx.Bind(&req); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if req.OrderID <= 0 {
		return BadRequest(ctx, "订单ID无效")
	}
	if req.ProjectID <= 0 {
		return BadRequest(ctx, "项目ID无效")
	}

	// 验证订单归属和状态
	order, err := s.repo.Order.GetByID(ctx.Request().Context(), req.OrderID)
	if err != nil {
		return InternalError(ctx, "获取订单失败")
	}
	if order == nil {
		return NotFound(ctx, "订单不存在")
	}
	if order.UserID != userID {
		return Forbidden(ctx, "无权操作此订单")
	}
	if order.Status != 1 {
		return BadRequest(ctx, "订单未支付或状态异常")
	}

	// 验证项目归属
	project, err := s.repo.Project.GetByID(ctx.Request().Context(), req.ProjectID)
	if err != nil {
		return InternalError(ctx, "获取项目失败")
	}
	if project == nil {
		return NotFound(ctx, "项目不存在")
	}
	if project.CreatorID != userID {
		return Forbidden(ctx, "只能推广自己创建的项目")
	}

	// 检查是否已经为此订单创建过推广
	existingPromotion, err := s.repo.EmailPromotion.GetByOrderID(ctx.Request().Context(), req.OrderID)
	if err != nil {
		return InternalError(ctx, "检查推广记录失败")
	}
	if existingPromotion != nil {
		return BadRequest(ctx, "此订单已触发过推广")
	}

	// 获取订单中的推广商品数量作为最大发送人数
	var maxRecipients int
	for _, item := range order.Items {
		product, err := s.repo.Product.GetByID(ctx.Request().Context(), item.ProductID)
		if err != nil || product == nil {
			continue
		}
		if product.Type == 2 { // 服务权益 - 邮件推广
			maxRecipients += item.Quantity
		}
	}

	if maxRecipients <= 0 {
		return BadRequest(ctx, "订单中没有邮件推广商品")
	}

	// 创建推广记录
	promotion := &models.EmailPromotion{
		OrderID:       req.OrderID,
		ProjectID:     req.ProjectID,
		CreatorID:     userID,
		MaxRecipients: maxRecipients,
		Status:        models.EmailPromotionStatusPending,
	}

	if err := s.repo.EmailPromotion.Create(ctx.Request().Context(), promotion); err != nil {
		ctx.Logger().Error("Failed to create email promotion: ", err)
		return InternalError(ctx, "创建推广记录失败")
	}

	// 尝试创建邮件服务并异步发送
	go func() {
		emailService, err := email.NewServiceFromEnv(
			s.repo.User,
			s.repo.Project,
			s.repo.EmailPromotion,
		)
		if err != nil {
			// 邮件服务未配置，记录错误但不影响推广记录创建
			errMsg := "邮件服务未配置: " + err.Error()
			promotion.Status = models.EmailPromotionStatusFailed
			promotion.ErrorMessage = &errMsg
			s.repo.EmailPromotion.Update(context.Background(), promotion)
			return
		}

		emailService.SendPromotionEmails(context.Background(), promotion)
	}()

	return Success(ctx, map[string]interface{}{
		"promotionId":   promotion.ID,
		"maxRecipients": maxRecipients,
		"status":        "pending",
		"message":       "推广任务已创建，正在发送中",
	})
}

// GetEmailPromotionStatus 获取推广状态
// GET /api/email/promotion/:id/status
func (s *Server) GetEmailPromotionStatus(ctx echo.Context, id int) error {
	userID := GetUserID(ctx)

	promotion, err := s.repo.EmailPromotion.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return InternalError(ctx, "获取推广记录失败")
	}
	if promotion == nil {
		return NotFound(ctx, "推广记录不存在")
	}
	if promotion.CreatorID != userID {
		return Forbidden(ctx, "无权查看此推广记录")
	}

	statusText := map[models.EmailPromotionStatus]string{
		models.EmailPromotionStatusPending:   "待发送",
		models.EmailPromotionStatusSending:   "发送中",
		models.EmailPromotionStatusCompleted: "已完成",
		models.EmailPromotionStatusFailed:    "发送失败",
	}

	return Success(ctx, map[string]interface{}{
		"id":            promotion.ID,
		"projectId":     promotion.ProjectID,
		"projectName":   promotion.ProjectName,
		"maxRecipients": promotion.MaxRecipients,
		"totalSent":     promotion.TotalSent,
		"status":        promotion.Status,
		"statusText":    statusText[promotion.Status],
		"errorMessage":  promotion.ErrorMessage,
		"startedAt":     promotion.StartedAt,
		"completedAt":   promotion.CompletedAt,
		"createdAt":     promotion.CreatedAt,
	})
}

// ListMyEmailPromotions 获取我的推广记录
// GET /api/email/promotions/my
func (s *Server) ListMyEmailPromotions(ctx echo.Context) error {
	userID := GetUserID(ctx)

	// 简单分页
	page := 1
	size := 10

	promotions, total, err := s.repo.EmailPromotion.ListByCreatorID(ctx.Request().Context(), userID, page, size)
	if err != nil {
		return InternalError(ctx, "获取推广记录失败")
	}

	statusText := map[models.EmailPromotionStatus]string{
		models.EmailPromotionStatusPending:   "待发送",
		models.EmailPromotionStatusSending:   "发送中",
		models.EmailPromotionStatusCompleted: "已完成",
		models.EmailPromotionStatusFailed:    "发送失败",
	}

	list := make([]map[string]interface{}, len(promotions))
	for i, p := range promotions {
		list[i] = map[string]interface{}{
			"id":            p.ID,
			"projectId":     p.ProjectID,
			"projectName":   p.ProjectName,
			"maxRecipients": p.MaxRecipients,
			"totalSent":     p.TotalSent,
			"status":        p.Status,
			"statusText":    statusText[p.Status],
			"createdAt":     p.CreatedAt,
		}
	}

	return Success(ctx, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

// getBaseURL 获取基础URL
func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://kuaizu.com"
	}
	return baseURL
}
