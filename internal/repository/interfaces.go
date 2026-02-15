package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// OrderRepo defines the interface for order repository operations used by services.
type OrderRepo interface {
	GetByID(ctx context.Context, id int) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	ListByUserID(ctx context.Context, params OrderListParams) ([]*models.Order, int64, error)
	UpdatePaymentStatus(ctx context.Context, id int, status int, wxPayNo string, payTime time.Time) error
	UpdatePaymentStatusTx(ctx context.Context, tx *sqlx.Tx, id int, status int, wxPayNo string, payTime time.Time) error
	GetOrderItems(ctx context.Context, orderID int) ([]*models.OrderItem, error)
}

// ProjectRepo defines the interface for project repository operations used by services.
type ProjectRepo interface {
	GetByID(ctx context.Context, id int) (*models.Project, error)
	List(ctx context.Context, params ListParams) ([]models.Project, int64, error)
	Create(ctx context.Context, p *models.Project) error
	Update(ctx context.Context, p *models.Project) error
	Delete(ctx context.Context, id int) error
	IsOwner(ctx context.Context, projectID, userID int) (bool, error)
	UpdateStatus(ctx context.Context, id int, status int) error
	IncrementViewCount(ctx context.Context, id int) error
}

// ProductRepo defines the interface for product repository operations used by services.
type ProductRepo interface {
	GetByID(ctx context.Context, id int) (*models.Product, error)
	GetAll(ctx context.Context) ([]*models.Product, error)
}

// EmailPromotionRepo defines the interface for email promotion repository operations.
type EmailPromotionRepo interface {
	Create(ctx context.Context, promotion *models.EmailPromotion) error
	GetByID(ctx context.Context, id int) (*models.EmailPromotion, error)
	GetByOrderID(ctx context.Context, orderID int) (*models.EmailPromotion, error)
	Update(ctx context.Context, promotion *models.EmailPromotion) error
	ListByCreatorID(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error)
	ListByProjectID(ctx context.Context, projectID int) ([]models.EmailPromotion, error)
}

// UserRepo defines the interface for user repository operations used by services.
type UserRepo interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByOpenID(ctx context.Context, openid string) (*models.User, error)
	Create(ctx context.Context, openid string) (*models.User, error)
	CreateWithPhone(ctx context.Context, openid string, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePhone(ctx context.Context, userID int, phone string) error
	UpdateQuota(ctx context.Context, user *models.User) error
	AddOliveBranchCount(ctx context.Context, userID int, count int) error
	AddOliveBranchCountTx(ctx context.Context, tx *sqlx.Tx, userID int, count int) error
	UpdateAuthStatus(ctx context.Context, userID int, authStatus int) error
	ListUsers(ctx context.Context, params UserListParams) ([]models.User, int64, error)
	FindEmailRecipients(ctx context.Context, excludeUserID int, limit int) ([]*EmailRecipient, error)
	SetEmailOptOut(ctx context.Context, userID int, optOut bool) error
}

// Compile-time interface satisfaction checks
var _ OrderRepo = (*OrderRepository)(nil)
var _ ProjectRepo = (*ProjectRepository)(nil)
var _ ProductRepo = (*ProductRepository)(nil)
var _ EmailPromotionRepo = (*EmailPromotionRepository)(nil)
var _ UserRepo = (*UserRepository)(nil)
