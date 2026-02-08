package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// EmailPromotionRepository handles email promotion database operations
type EmailPromotionRepository struct {
	db *sqlx.DB
}

// NewEmailPromotionRepository creates a new EmailPromotionRepository
func NewEmailPromotionRepository(db *sqlx.DB) *EmailPromotionRepository {
	return &EmailPromotionRepository{db: db}
}

// Create creates a new email promotion record
func (r *EmailPromotionRepository) Create(ctx context.Context, promotion *models.EmailPromotion) error {
	query := `
		INSERT INTO email_promotion (order_id, project_id, creator_id, max_recipients, total_sent, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		promotion.OrderID,
		promotion.ProjectID,
		promotion.CreatorID,
		promotion.MaxRecipients,
		promotion.TotalSent,
		promotion.Status,
	)
	if err != nil {
		return fmt.Errorf("create email promotion: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	promotion.ID = int(id)
	return nil
}

// GetByID retrieves an email promotion by ID
func (r *EmailPromotionRepository) GetByID(ctx context.Context, id int) (*models.EmailPromotion, error) {
	query := `
		SELECT 
			ep.id, ep.order_id, ep.project_id, ep.creator_id,
			ep.max_recipients, ep.total_sent, ep.status,
			ep.error_message, ep.started_at, ep.completed_at, ep.created_at,
			p.name AS project_name
		FROM email_promotion ep
		LEFT JOIN project p ON ep.project_id = p.id
		WHERE ep.id = ?
	`

	var promotion models.EmailPromotion
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&promotion.ID, &promotion.OrderID, &promotion.ProjectID, &promotion.CreatorID,
		&promotion.MaxRecipients, &promotion.TotalSent, &promotion.Status,
		&promotion.ErrorMessage, &promotion.StartedAt, &promotion.CompletedAt, &promotion.CreatedAt,
		&promotion.ProjectName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get email promotion by id: %w", err)
	}

	return &promotion, nil
}

// GetByOrderID retrieves an email promotion by order ID
func (r *EmailPromotionRepository) GetByOrderID(ctx context.Context, orderID int) (*models.EmailPromotion, error) {
	query := `
		SELECT 
			id, order_id, project_id, creator_id,
			max_recipients, total_sent, status,
			error_message, started_at, completed_at, created_at
		FROM email_promotion
		WHERE order_id = ?
	`

	var promotion models.EmailPromotion
	err := r.db.QueryRowxContext(ctx, query, orderID).Scan(
		&promotion.ID, &promotion.OrderID, &promotion.ProjectID, &promotion.CreatorID,
		&promotion.MaxRecipients, &promotion.TotalSent, &promotion.Status,
		&promotion.ErrorMessage, &promotion.StartedAt, &promotion.CompletedAt, &promotion.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get email promotion by order id: %w", err)
	}

	return &promotion, nil
}

// Update updates an email promotion record
func (r *EmailPromotionRepository) Update(ctx context.Context, promotion *models.EmailPromotion) error {
	query := `
		UPDATE email_promotion SET
			total_sent = ?,
			status = ?,
			error_message = ?,
			started_at = ?,
			completed_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		promotion.TotalSent,
		promotion.Status,
		promotion.ErrorMessage,
		promotion.StartedAt,
		promotion.CompletedAt,
		promotion.ID,
	)
	if err != nil {
		return fmt.Errorf("update email promotion: %w", err)
	}

	return nil
}

// ListByCreatorID retrieves email promotions by creator ID with pagination
func (r *EmailPromotionRepository) ListByCreatorID(ctx context.Context, creatorID int, page, size int) ([]models.EmailPromotion, int64, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM email_promotion WHERE creator_id = ?`
	var total int64
	if err := r.db.QueryRowxContext(ctx, countQuery, creatorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count email promotions: %w", err)
	}

	// Query with pagination
	offset := (page - 1) * size
	query := `
		SELECT 
			ep.id, ep.order_id, ep.project_id, ep.creator_id,
			ep.max_recipients, ep.total_sent, ep.status,
			ep.error_message, ep.started_at, ep.completed_at, ep.created_at,
			p.name AS project_name
		FROM email_promotion ep
		LEFT JOIN project p ON ep.project_id = p.id
		WHERE ep.creator_id = ?
		ORDER BY ep.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryxContext(ctx, query, creatorID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query email promotions: %w", err)
	}
	defer rows.Close()

	var promotions []models.EmailPromotion
	for rows.Next() {
		var promotion models.EmailPromotion
		err := rows.Scan(
			&promotion.ID, &promotion.OrderID, &promotion.ProjectID, &promotion.CreatorID,
			&promotion.MaxRecipients, &promotion.TotalSent, &promotion.Status,
			&promotion.ErrorMessage, &promotion.StartedAt, &promotion.CompletedAt, &promotion.CreatedAt,
			&promotion.ProjectName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan email promotion: %w", err)
		}
		promotions = append(promotions, promotion)
	}

	return promotions, total, nil
}

// ListByProjectID retrieves email promotions by project ID
func (r *EmailPromotionRepository) ListByProjectID(ctx context.Context, projectID int) ([]models.EmailPromotion, error) {
	query := `
		SELECT 
			id, order_id, project_id, creator_id,
			max_recipients, total_sent, status,
			error_message, started_at, completed_at, created_at
		FROM email_promotion
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query email promotions by project: %w", err)
	}
	defer rows.Close()

	var promotions []models.EmailPromotion
	for rows.Next() {
		var promotion models.EmailPromotion
		err := rows.Scan(
			&promotion.ID, &promotion.OrderID, &promotion.ProjectID, &promotion.CreatorID,
			&promotion.MaxRecipients, &promotion.TotalSent, &promotion.Status,
			&promotion.ErrorMessage, &promotion.StartedAt, &promotion.CompletedAt, &promotion.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan email promotion: %w", err)
		}
		promotions = append(promotions, promotion)
	}

	return promotions, nil
}
