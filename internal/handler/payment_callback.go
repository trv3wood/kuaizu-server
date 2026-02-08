package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/wechat"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

type wechatPayResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func successResponse() wechatPayResponse {
	return wechatPayResponse{Code: "SUCCESS", Message: "成功"}
}

func failResponse(message string) wechatPayResponse {
	return wechatPayResponse{Code: "FAIL", Message: message}
}

// initWechatPayClient initializes and returns a WeChat Pay client
func (s *Server) initWechatPayClient() (*wechat.PayClient, error) {
	payConfig, err := wechat.DefaultPayConfig()
	if err != nil {
		return nil, err
	}

	return wechat.NewPayClient(payConfig)
}

// parseAndVerifyNotification parses and verifies the WeChat Pay notification
func (s *Server) parseAndVerifyNotification(ctx echo.Context, payClient *wechat.PayClient) (*payments.Transaction, error) {
	return payClient.ParseNotification(ctx.Request().Context(), ctx.Request())
}

// validateTransaction validates the transaction and extracts order ID
func (s *Server) validateTransaction(ctx echo.Context, transaction *payments.Transaction) (int, bool) {
	if transaction.OutTradeNo == nil {
		return 0, false
	}

	orderID, err := wechat.ParseOrderIDFromOutTradeNo(*transaction.OutTradeNo)
	if err != nil {
		return 0, false
	}

	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		s.repo.Order.UpdatePaymentStatus(ctx.Request().Context(), orderID, 2, "", time.Now())
		return orderID, false
	}

	return orderID, true
}

// validateOrder retrieves and validates the order
func (s *Server) validateOrder(ctx context.Context, orderID int) (*models.Order, error) {
	order, err := s.repo.Order.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}
	return order, nil
}

// extractPaymentInfo extracts payment time and transaction ID from transaction
func (s *Server) extractPaymentInfo(transaction *payments.Transaction) (time.Time, string) {
	var payTime time.Time
	if transaction.SuccessTime != nil {
		payTime, _ = time.Parse(time.RFC3339, *transaction.SuccessTime)
	}
	if payTime.IsZero() {
		payTime = time.Now()
	}

	transactionID := ""
	if transaction.TransactionId != nil {
		transactionID = *transaction.TransactionId
	}

	return payTime, transactionID
}

// processPayment updates order status and distributes benefits within a transaction
func (s *Server) processPayment(ctx echo.Context, order *models.Order, transactionID string, payTime time.Time) error {
	tx, err := s.repo.DB().BeginTxx(ctx.Request().Context(), nil)
	if err != nil {
		ctx.Logger().Error("Failed to begin transaction: ", err)
		return err
	}
	defer tx.Rollback()

	if err := s.updateOrderStatus(ctx, tx, order.ID, transactionID, payTime); err != nil {
		return err
	}

	if err := s.distributeBenefits(ctx, tx, order); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		ctx.Logger().Error("Failed to commit transaction: ", err)
		return err
	}

	return nil
}

// updateOrderStatus updates the order payment status
func (s *Server) updateOrderStatus(ctx echo.Context, tx *sqlx.Tx, orderID int, transactionID string, payTime time.Time) error {
	err := s.repo.Order.UpdatePaymentStatusTx(
		ctx.Request().Context(),
		tx,
		orderID,
		1,
		transactionID,
		payTime,
	)
	if err != nil {
		ctx.Logger().Error("Failed to update order status: ", err)
		return err
	}
	return nil
}

// distributeBenefits distributes benefits based on product types in the order
func (s *Server) distributeBenefits(ctx echo.Context, tx *sqlx.Tx, order *models.Order) error {
	for _, item := range order.Items {
		product, err := s.repo.Product.GetByID(ctx.Request().Context(), item.ProductID)
		if err != nil || product == nil {
			continue
		}

		switch product.Type {
		case 1:
			if err := s.repo.User.AddOliveBranchCountTx(ctx.Request().Context(), tx, order.UserID, item.Quantity); err != nil {
				ctx.Logger().Error("Failed to add olive branch count: ", err)
				return err
			}
		case 2:
		default:
			ctx.Logger().Error("Unknown product type: ", product.Type)
		}
	}
	return nil
}

// WechatPayCallback handles POST /payment/wechat/notify
func (s *Server) WechatPayCallback(ctx echo.Context) error {
	payClient, err := s.initWechatPayClient()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, failResponse("支付配置错误"))
	}

	transaction, err := s.parseAndVerifyNotification(ctx, payClient)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, failResponse("验签失败: "+err.Error()))
	}

	orderID, isSuccess := s.validateTransaction(ctx, transaction)
	if !isSuccess {
		return ctx.JSON(http.StatusOK, successResponse())
	}

	order, err := s.validateOrder(ctx.Request().Context(), orderID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, failResponse("查询订单失败"))
	}
	if order == nil {
		return ctx.JSON(http.StatusOK, successResponse())
	}

	if order.Status == 1 {
		return ctx.JSON(http.StatusOK, successResponse())
	}

	payTime, transactionID := s.extractPaymentInfo(transaction)

	if err := s.processPayment(ctx, order, transactionID, payTime); err != nil {
		return ctx.JSON(http.StatusInternalServerError, failResponse("处理支付失败"))
	}

	return ctx.JSON(http.StatusOK, successResponse())
}
