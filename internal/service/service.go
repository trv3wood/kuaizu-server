package service

import "github.com/trv3wood/kuaizu-server/internal/repository"

// Services aggregates all service instances.
type Services struct {
	EmailPromotion   *EmailPromotionService
	Payment          *PaymentService
	EmailUnsubscribe *EmailUnsubscribeService
	Order            *OrderService
	OliveBranch      *OliveBranchService
}

// New creates a new Services instance with all sub-services.
func New(repo *repository.Repository) *Services {
	return &Services{
		EmailPromotion:   NewEmailPromotionService(repo),
		Payment:          NewPaymentService(repo),
		EmailUnsubscribe: NewEmailUnsubscribeService(repo),
		Order:            NewOrderService(repo),
		OliveBranch:      NewOliveBranchService(repo),
	}
}
