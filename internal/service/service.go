package service

import (
	"github.com/trv3wood/kuaizu-server/internal/oss"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// Services aggregates all service instances.
type Services struct {
	Auth             *AuthService
	EmailPromotion   *EmailPromotionService
	Payment          *PaymentService
	EmailUnsubscribe *EmailUnsubscribeService
	Order            *OrderService
	OliveBranch      *OliveBranchService
	Commons          *CommonsService
}

// New creates a new Services instance with all sub-services.
func New(repo *repository.Repository, ossClient *oss.Client) *Services {
	return &Services{
		Auth:             NewAuthService(repo),
		EmailPromotion:   NewEmailPromotionService(repo),
		Payment:          NewPaymentService(repo),
		EmailUnsubscribe: NewEmailUnsubscribeService(repo),
		Order:            NewOrderService(repo),
		OliveBranch:      NewOliveBranchService(repo),
		Commons:          NewCommonsService(ossClient),
	}
}
