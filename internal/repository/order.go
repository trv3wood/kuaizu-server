package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// OrderRepository handles order database operations
type OrderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// OrderListParams contains parameters for listing orders
type OrderListParams struct {
	UserID int
	Page   int
	Size   int
	Status *int
}

// ListByUserID retrieves paginated orders for a user
func (r *OrderRepository) ListByUserID(ctx context.Context, params OrderListParams) ([]*models.Order, int64, error) {
	// Build where clause
	where := `WHERE o.user_id = ?`
	args := []interface{}{params.UserID}

	if params.Status != nil {
		where += ` AND o.status = ?`
		args = append(args, *params.Status)
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `order` o %s", where)
	var total int64
	err := r.db.QueryRowxContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	// Query with pagination
	offset := (params.Page - 1) * params.Size
	query := fmt.Sprintf(`
		SELECT
			o.id, o.user_id, o.product_id, o.actual_paid, o.status,
			o.wx_pay_no, o.pay_time, o.created_at, o.updated_at,
			p.name as product_name
		FROM `+"`order`"+` o
		LEFT JOIN product p ON o.product_id = p.id
		%s
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	args = append(args, params.Size, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.ID, &o.UserID, &o.ProductID, &o.ActualPaid, &o.Status,
			&o.WxPayNo, &o.PayTime, &o.CreatedAt, &o.UpdatedAt,
			&o.ProductName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, &o)
	}

	return orders, total, nil
}
