package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trv3wood/kuaizu-server/internal/models"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// --- Mock Repositories ---

// mockOrderRepo is a mock implementation of repository.OrderRepo.
type mockOrderRepo struct {
	getByIDFn func(ctx context.Context, id int) (*models.Order, error)
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int) (*models.Order, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockOrderRepo) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) ListByUserID(ctx context.Context, params repository.OrderListParams) ([]*models.Order, int64, error) {
	return nil, 0, nil
}
func (m *mockOrderRepo) UpdatePaymentStatus(ctx context.Context, id int, status int, wxPayNo string, payTime time.Time) error {
	return nil
}
func (m *mockOrderRepo) UpdatePaymentStatusTx(ctx context.Context, tx *sqlx.Tx, id int, status int, wxPayNo string, payTime time.Time) error {
	return nil
}
func (m *mockOrderRepo) GetOrderItems(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	return nil, nil
}

// mockProjectRepo is a mock implementation of repository.ProjectRepo.
type mockProjectRepo struct {
	getByIDFn func(ctx context.Context, id int) (*models.Project, error)
}

func (m *mockProjectRepo) GetByID(ctx context.Context, id int) (*models.Project, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockProjectRepo) List(ctx context.Context, params repository.ListParams) ([]models.Project, int64, error) {
	return nil, 0, nil
}
func (m *mockProjectRepo) Create(ctx context.Context, p *models.Project) error { return nil }
func (m *mockProjectRepo) Update(ctx context.Context, p *models.Project) error { return nil }
func (m *mockProjectRepo) Delete(ctx context.Context, id int) error            { return nil }
func (m *mockProjectRepo) IsOwner(ctx context.Context, projectID, userID int) (bool, error) {
	return false, nil
}
func (m *mockProjectRepo) UpdateStatus(ctx context.Context, id int, status int) error { return nil }
func (m *mockProjectRepo) IncrementViewCount(ctx context.Context, id int) error       { return nil }

// mockProductRepo is a mock implementation of repository.ProductRepo.
type mockProductRepo struct {
	getByIDFn func(ctx context.Context, id int) (*models.Product, error)
}

func (m *mockProductRepo) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockProductRepo) GetAll(ctx context.Context) ([]*models.Product, error) {
	return nil, nil
}

// mockEmailPromotionRepo is a mock implementation of repository.EmailPromotionRepo.
type mockEmailPromotionRepo struct {
	createFn          func(ctx context.Context, promotion *models.EmailPromotion) error
	getByIDFn         func(ctx context.Context, id int) (*models.EmailPromotion, error)
	getByOrderIDFn    func(ctx context.Context, orderID int) (*models.EmailPromotion, error)
	updateFn          func(ctx context.Context, promotion *models.EmailPromotion) error
	listByCreatorIDFn func(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error)
}

func (m *mockEmailPromotionRepo) Create(ctx context.Context, promotion *models.EmailPromotion) error {
	if m.createFn != nil {
		return m.createFn(ctx, promotion)
	}
	return nil
}
func (m *mockEmailPromotionRepo) GetByID(ctx context.Context, id int) (*models.EmailPromotion, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockEmailPromotionRepo) GetByOrderID(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
	if m.getByOrderIDFn != nil {
		return m.getByOrderIDFn(ctx, orderID)
	}
	return nil, nil
}
func (m *mockEmailPromotionRepo) Update(ctx context.Context, promotion *models.EmailPromotion) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, promotion)
	}
	return nil
}
func (m *mockEmailPromotionRepo) ListByCreatorID(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error) {
	if m.listByCreatorIDFn != nil {
		return m.listByCreatorIDFn(ctx, creatorID, page, size)
	}
	return nil, 0, nil
}
func (m *mockEmailPromotionRepo) ListByProjectID(ctx context.Context, projectID int) ([]models.EmailPromotion, error) {
	return nil, nil
}

// --- Helper to create repo with mocks ---

func newMockRepo(
	orderRepo *mockOrderRepo,
	projectRepo *mockProjectRepo,
	productRepo *mockProductRepo,
	emailPromotionRepo *mockEmailPromotionRepo,
) *repository.Repository {
	return &repository.Repository{
		Order:          orderRepo,
		Project:        projectRepo,
		Product:        productRepo,
		EmailPromotion: emailPromotionRepo,
	}
}

// --- Tests for TriggerPromotion ---

func TestTriggerPromotion_OrderNotFound(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return nil, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeNotFound, "订单不存在")
}

func TestTriggerPromotion_OrderGetError(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return nil, errors.New("db error")
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeInternal, "获取订单失败")
}

func TestTriggerPromotion_OrderNotOwnedByUser(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{ID: 100, UserID: 999, Status: 1}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeForbidden, "无权操作此订单")
}

func TestTriggerPromotion_OrderNotPaid(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{ID: 100, UserID: 1, Status: 0}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeBadRequest, "订单未支付或状态异常")
}

func TestTriggerPromotion_ProjectNotFound(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{ID: 100, UserID: 1, Status: 1}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return nil, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeNotFound, "项目不存在")
}

func TestTriggerPromotion_ProjectNotOwnedByUser(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{ID: 100, UserID: 1, Status: 1}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return &models.Project{ID: 200, CreatorID: 999}, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeForbidden, "只能推广自己创建的项目")
}

func TestTriggerPromotion_AlreadyTriggered(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{ID: 100, UserID: 1, Status: 1}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return &models.Project{ID: 200, CreatorID: 1}, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			getByOrderIDFn: func(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
				return &models.EmailPromotion{ID: 1, OrderID: orderID}, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeBadRequest, "此订单已触发过推广")
}

func TestTriggerPromotion_NoEmailPromotionProduct(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{
				ID: 100, UserID: 1, Status: 1,
				Items: []*models.OrderItem{{ProductID: 10, Quantity: 1}},
			}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return &models.Project{ID: 200, CreatorID: 1}, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			return &models.Product{ID: 10, Type: 1}, nil // Type 1 = 虚拟币, not email
		}},
		&mockEmailPromotionRepo{
			getByOrderIDFn: func(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
				return nil, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeBadRequest, "订单中没有邮件推广商品")
}

func TestTriggerPromotion_Success(t *testing.T) {
	var createdPromotion *models.EmailPromotion

	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{
				ID: 100, UserID: 1, Status: 1,
				Items: []*models.OrderItem{
					{ProductID: 10, Quantity: 50},
					{ProductID: 11, Quantity: 30},
				},
			}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return &models.Project{ID: 200, CreatorID: 1}, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			if id == 10 {
				return &models.Product{ID: 10, Type: 2}, nil
			}
			if id == 11 {
				return &models.Product{ID: 11, Type: 2}, nil
			}
			return nil, nil
		}},
		&mockEmailPromotionRepo{
			getByOrderIDFn: func(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, promotion *models.EmailPromotion) error {
				promotion.ID = 1
				createdPromotion = promotion
				return nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	result, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 80, result.MaxRecipients, "50 + 30 = 80 recipients")
	assert.Equal(t, 100, result.Promotion.OrderID)
	assert.Equal(t, 200, result.Promotion.ProjectID)
	assert.Equal(t, 1, result.Promotion.CreatorID)
	assert.Equal(t, models.EmailPromotionStatusPending, result.Promotion.Status)

	require.NotNil(t, createdPromotion, "promotion should be created via repo")
}

func TestTriggerPromotion_CreateFails(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) {
			return &models.Order{
				ID: 100, UserID: 1, Status: 1,
				Items: []*models.OrderItem{{ProductID: 10, Quantity: 50}},
			}, nil
		}},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) {
			return &models.Project{ID: 200, CreatorID: 1}, nil
		}},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			return &models.Product{ID: 10, Type: 2}, nil
		}},
		&mockEmailPromotionRepo{
			getByOrderIDFn: func(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, promotion *models.EmailPromotion) error {
				return errors.New("db write error")
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.TriggerPromotion(context.Background(), 1, 100, 200)

	assertServiceError(t, err, ErrCodeInternal, "创建推广记录失败")
}

// --- Tests for GetStatus ---

func TestGetStatus_NotFound(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			getByIDFn: func(ctx context.Context, id int) (*models.EmailPromotion, error) {
				return nil, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.GetStatus(context.Background(), 1, 999)

	assertServiceError(t, err, ErrCodeNotFound, "推广记录不存在")
}

func TestGetStatus_Forbidden(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			getByIDFn: func(ctx context.Context, id int) (*models.EmailPromotion, error) {
				return &models.EmailPromotion{ID: 1, CreatorID: 999}, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.GetStatus(context.Background(), 1, 1)

	assertServiceError(t, err, ErrCodeForbidden, "无权查看此推广记录")
}

func TestGetStatus_Success(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			getByIDFn: func(ctx context.Context, id int) (*models.EmailPromotion, error) {
				return &models.EmailPromotion{
					ID:            1,
					CreatorID:     1,
					OrderID:       100,
					ProjectID:     200,
					MaxRecipients: 50,
					Status:        models.EmailPromotionStatusCompleted,
				}, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	promotion, err := svc.GetStatus(context.Background(), 1, 1)

	require.NoError(t, err)
	assert.Equal(t, 1, promotion.ID)
	assert.Equal(t, models.EmailPromotionStatusCompleted, promotion.Status)
}

func TestGetStatus_RepoError(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			getByIDFn: func(ctx context.Context, id int) (*models.EmailPromotion, error) {
				return nil, errors.New("db error")
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, err := svc.GetStatus(context.Background(), 1, 1)

	assertServiceError(t, err, ErrCodeInternal, "获取推广记录失败")
}

// --- Tests for ListByCreator ---

func TestListByCreator_Success(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			listByCreatorIDFn: func(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error) {
				return []models.EmailPromotion{
					{ID: 1, CreatorID: creatorID},
					{ID: 2, CreatorID: creatorID},
				}, 2, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	promotions, total, err := svc.ListByCreator(context.Background(), 1, 1, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, promotions, 2)
}

func TestListByCreator_RepoError(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			listByCreatorIDFn: func(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error) {
				return nil, 0, errors.New("db error")
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	_, _, err := svc.ListByCreator(context.Background(), 1, 1, 10)

	assertServiceError(t, err, ErrCodeInternal, "获取推广记录失败")
}

func TestListByCreator_Empty(t *testing.T) {
	repo := newMockRepo(
		&mockOrderRepo{getByIDFn: func(ctx context.Context, id int) (*models.Order, error) { return nil, nil }},
		&mockProjectRepo{getByIDFn: func(ctx context.Context, id int) (*models.Project, error) { return nil, nil }},
		&mockProductRepo{getByIDFn: func(ctx context.Context, id int) (*models.Product, error) { return nil, nil }},
		&mockEmailPromotionRepo{
			listByCreatorIDFn: func(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error) {
				return nil, 0, nil
			},
		},
	)

	svc := NewEmailPromotionService(repo)
	promotions, total, err := svc.ListByCreator(context.Background(), 1, 1, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, promotions)
}

// --- Test Helper ---

func assertServiceError(t *testing.T, err error, expectedCode ErrorCode, expectedMsg string) {
	t.Helper()
	require.Error(t, err)
	svcErr, ok := err.(*ServiceError)
	require.True(t, ok, "expected *ServiceError, got %T: %v", err, err)
	assert.Equal(t, expectedCode, svcErr.Code)
	assert.Equal(t, expectedMsg, svcErr.Message)
}
