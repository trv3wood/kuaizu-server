package service

import (
	"context"
	"log"
	"time"

	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// PaymentService handles payment-related business logic.
type PaymentService struct {
	repo *repository.Repository
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo *repository.Repository) *PaymentService {
	return &PaymentService{repo: repo}
}

// GetOrder retrieves an order by ID (returns nil, nil if not found).
func (s *PaymentService) GetOrder(ctx context.Context, orderID int) (*models.Order, error) {
	order, err := s.repo.Order.GetByID(ctx, orderID)
	if err != nil {
		log.Printf("[PaymentService.GetOrder] repository error: %v", err)
		return nil, ErrInternal("查询订单失败")
	}
	return order, nil
}

// MarkPaymentFailed updates order status to failed.
func (s *PaymentService) MarkPaymentFailed(ctx context.Context, orderID int) {
	s.repo.Order.UpdatePaymentStatus(ctx, orderID, 2, "", time.Now())
}

// ProcessPayment updates order status and distributes benefits within a DB transaction.
func (s *PaymentService) ProcessPayment(ctx context.Context, order *models.Order, transactionID string, payTime time.Time) error {
	tx, err := s.repo.DB().BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[PaymentService.ProcessPayment] failed to begin transaction: %v", err)
		return ErrInternal("处理支付失败")
	}
	defer tx.Rollback()

	// Update order status
	if err := s.repo.Order.UpdatePaymentStatusTx(ctx, tx, order.ID, 1, transactionID, payTime); err != nil {
		log.Printf("[PaymentService.ProcessPayment] failed to update order status: %v", err)
		return ErrInternal("处理支付失败")
	}

	// Distribute benefits
	product, err := s.repo.Product.GetByID(ctx, order.ProductID)
	if err != nil || product == nil {
		log.Printf("[PaymentService.ProcessPayment] failed to get product: %v", err)
		return ErrInternal("处理支付失败")
	}

	switch product.Type {
	case 1: // 橄榄枝
		if err := s.repo.User.AddOliveBranchCountTx(ctx, tx, order.UserID, order.Quantity); err != nil {
			log.Printf("[PaymentService.ProcessPayment] failed to add olive branch count: %v", err)
			return ErrInternal("处理支付失败")
		}
	case 2:
		// 权益需要凭订单和参数手动兑换
	default:
		log.Printf("[PaymentService.ProcessPayment] unknown product type: %d", product.Type)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[PaymentService.ProcessPayment] failed to commit transaction: %v", err)
		return ErrInternal("处理支付失败")
	}

	return nil
}
