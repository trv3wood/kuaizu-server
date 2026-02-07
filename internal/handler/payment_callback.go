package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/internal/wechat"
)

// WechatPayCallback handles POST /payment/wechat/notify
func (s *Server) WechatPayCallback(ctx echo.Context) error {
	// Get WeChat Pay config
	payConfig, err := wechat.DefaultPayConfig()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "支付配置错误",
		})
	}

	// Create pay client
	payClient, err := wechat.NewPayClient(payConfig)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "初始化支付客户端失败",
		})
	}

	// Parse and verify notification using SDK
	transaction, err := payClient.ParseNotification(ctx.Request().Context(), ctx.Request())
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"code":    "FAIL",
			"message": "验签失败: " + err.Error(),
		})
	}

	// Parse order ID from out_trade_no
	orderID, err := wechat.ParseOrderIDFromOutTradeNo(*transaction.OutTradeNo)

	// Check trade state
	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		// update order status to failed
		s.repo.Order.UpdatePaymentStatus(ctx.Request().Context(), orderID, 2, "", time.Now())
		// Payment not successful, just acknowledge
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    "SUCCESS",
			"message": "成功",
		})
	}

	if transaction.OutTradeNo == nil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    "SUCCESS",
			"message": "缺少订单号",
		})
	}

	if err != nil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    "SUCCESS",
			"message": "无效的订单号格式",
		})
	}

	// Find order by ID
	order, err := s.repo.Order.GetByID(ctx.Request().Context(), orderID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "查询订单失败",
		})
	}
	if order == nil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    "SUCCESS",
			"message": "订单不存在",
		})
	}

	// Already paid
	if order.Status == 1 {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    "SUCCESS",
			"message": "成功",
		})
	}

	// Parse success time
	var payTime time.Time
	if transaction.SuccessTime != nil {
		// SDK returns RFC3339 formatted string
		payTime, _ = time.Parse(time.RFC3339, *transaction.SuccessTime)
	}
	if payTime.IsZero() {
		payTime = time.Now()
	}

	// Get transaction ID
	transactionID := ""
	if transaction.TransactionId != nil {
		transactionID = *transaction.TransactionId
	}

	// Begin transaction for order status update and benefit distribution
	tx, err := s.repo.DB().BeginTxx(ctx.Request().Context(), nil)
	if err != nil {
		ctx.Logger().Error("Failed to begin transaction: ", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "开始事务失败",
		})
	}
	defer tx.Rollback()

	// Update order status within transaction
	err = s.repo.Order.UpdatePaymentStatusTx(
		ctx.Request().Context(),
		tx,
		order.ID,
		1, // 已支付
		transactionID,
		payTime,
	)
	if err != nil {
		ctx.Logger().Error("Failed to update order status: ", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "更新订单状态失败",
		})
	}

	// Process post-payment logic: distribute benefits based on product type
	for _, item := range order.Items {
		product, err := s.repo.Product.GetByID(ctx.Request().Context(), item.ProductID)
		if err != nil || product == nil {
			continue // Skip if product not found
		}

		switch product.Type {
		case 1: // 虚拟币 - 橄榄枝
			// Add olive branch count to user within transaction
			if err := s.repo.User.AddOliveBranchCountTx(ctx.Request().Context(), tx, order.UserID, item.Quantity); err != nil {
				ctx.Logger().Error("Failed to add olive branch count: ", err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"code":    "FAIL",
					"message": "分发权益失败",
				})
			}
		case 2: // 服务权益
			// TODO: Implement service benefit distribution (e.g., email promotion)
		default:
			ctx.Logger().Error("Unknown product type: ", product.Type)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		ctx.Logger().Error("Failed to commit transaction: ", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "FAIL",
			"message": "提交事务失败",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"code":    "SUCCESS",
		"message": "成功",
	})
}
