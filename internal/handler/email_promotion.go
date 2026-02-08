package handler

import (
	"context"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/email"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// TriggerEmailPromotion 触发邮件推广
// POST /api/email/promotion/trigger
// 用户购买邮件推广商品后，手动选择要推广的项目
func (s *Server) TriggerEmailPromotion(ctx echo.Context) error {
	userID := GetUserID(ctx)

	var body api.TriggerEmailPromotionJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return BadRequest(ctx, "请求参数错误")
	}

	if err := s.validatePromotionRequest(&body); err != nil {
		return err
	}

	order, err := s.validateOrderOwnership(ctx, body.OrderId, userID)
	if err != nil {
		return err
	}

	if err := s.validateProjectOwnership(ctx, body.ProjectId, userID); err != nil {
		return err
	}

	if err := s.checkPromotionDuplication(ctx, body.OrderId); err != nil {
		return err
	}

	maxRecipients, err := s.calculateMaxRecipients(ctx, order)
	if err != nil {
		return err
	}

	promotion, err := s.createPromotionRecord(ctx, body.OrderId, body.ProjectId, userID, maxRecipients)
	if err != nil {
		return err
	}

	s.startAsyncEmailSending(promotion)

	return s.buildPromotionResponse(ctx, promotion, maxRecipients)
}

func (s *Server) validatePromotionRequest(body *api.TriggerEmailPromotionJSONRequestBody) error {
	if body.OrderId <= 0 {
		return BadRequest(nil, "订单ID无效")
	}
	if body.ProjectId <= 0 {
		return BadRequest(nil, "项目ID无效")
	}
	return nil
}

func (s *Server) validateOrderOwnership(ctx echo.Context, orderID, userID int) (*models.Order, error) {
	order, err := s.repo.Order.GetByID(ctx.Request().Context(), orderID)
	if err != nil {
		return nil, InternalError(ctx, "获取订单失败")
	}
	if order == nil {
		return nil, NotFound(ctx, "订单不存在")
	}
	if order.UserID != userID {
		return nil, Forbidden(ctx, "无权操作此订单")
	}
	if order.Status != 1 {
		return nil, BadRequest(ctx, "订单未支付或状态异常")
	}
	return order, nil
}

func (s *Server) validateProjectOwnership(ctx echo.Context, projectID, userID int) error {
	project, err := s.repo.Project.GetByID(ctx.Request().Context(), projectID)
	if err != nil {
		return InternalError(ctx, "获取项目失败")
	}
	if project == nil {
		return NotFound(ctx, "项目不存在")
	}
	if project.CreatorID != userID {
		return Forbidden(ctx, "只能推广自己创建的项目")
	}
	return nil
}

func (s *Server) checkPromotionDuplication(ctx echo.Context, orderID int) error {
	existingPromotion, err := s.repo.EmailPromotion.GetByOrderID(ctx.Request().Context(), orderID)
	if err != nil {
		return InternalError(ctx, "检查推广记录失败")
	}
	if existingPromotion != nil {
		return BadRequest(ctx, "此订单已触发过推广")
	}
	return nil
}

func (s *Server) calculateMaxRecipients(ctx echo.Context, order *models.Order) (int, error) {
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
		return 0, BadRequest(ctx, "订单中没有邮件推广商品")
	}
	return maxRecipients, nil
}

func (s *Server) createPromotionRecord(ctx echo.Context, orderID, projectID, userID, maxRecipients int) (*models.EmailPromotion, error) {
	promotion := &models.EmailPromotion{
		OrderID:       orderID,
		ProjectID:     projectID,
		CreatorID:     userID,
		MaxRecipients: maxRecipients,
		Status:        models.EmailPromotionStatusPending,
	}

	if err := s.repo.EmailPromotion.Create(ctx.Request().Context(), promotion); err != nil {
		ctx.Logger().Error("Failed to create email promotion: ", err)
		return nil, InternalError(ctx, "创建推广记录失败")
	}
	return promotion, nil
}

func (s *Server) startAsyncEmailSending(promotion *models.EmailPromotion) {
	go func() {
		emailService, err := email.NewServiceFromEnv(
			s.repo.User,
			s.repo.Project,
			s.repo.EmailPromotion,
		)
		if err != nil {
			errMsg := "邮件服务未配置: " + err.Error()
			promotion.Status = models.EmailPromotionStatusFailed
			promotion.ErrorMessage = &errMsg
			s.repo.EmailPromotion.Update(context.Background(), promotion)
			return
		}

		emailService.SendPromotionEmails(context.Background(), promotion)
	}()
}

func (s *Server) buildPromotionResponse(ctx echo.Context, promotion *models.EmailPromotion, maxRecipients int) error {
	status := "pending"
	message := "推广任务已创建，正在发送中"

	return Success(ctx, api.TriggerEmailPromotionResponse{
		MaxRecipients: &maxRecipients,
		PromotionId:   &promotion.ID,
		Status:        &status,
		Message:       &message,
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

	statusTextValue := statusText[promotion.Status]
	statusValue := api.EmailPromotionStatus(promotion.Status)

	return Success(ctx, api.EmailPromotionVO{
		Id:            &promotion.ID,
		ProjectId:     &promotion.ProjectID,
		ProjectName:   promotion.ProjectName,
		MaxRecipients: &promotion.MaxRecipients,
		TotalSent:     &promotion.TotalSent,
		Status:        &statusValue,
		StatusText:    &statusTextValue,
		ErrorMessage:  promotion.ErrorMessage,
		StartedAt:     promotion.StartedAt,
		CompletedAt:   promotion.CompletedAt,
		CreatedAt:     &promotion.CreatedAt,
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

	list := make([]api.EmailPromotionVO, len(promotions))
	for i, p := range promotions {
		statusTextValue := statusText[p.Status]
		statusValue := api.EmailPromotionStatus(p.Status)

		list[i] = api.EmailPromotionVO{
			Id:            &p.ID,
			ProjectId:     &p.ProjectID,
			ProjectName:   p.ProjectName,
			MaxRecipients: &p.MaxRecipients,
			TotalSent:     &p.TotalSent,
			Status:        &statusValue,
			StatusText:    &statusTextValue,
			CreatedAt:     &p.CreatedAt,
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
