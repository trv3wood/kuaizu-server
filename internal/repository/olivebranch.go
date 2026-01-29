package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// OliveBranchRepository handles olive branch database operations
type OliveBranchRepository struct {
	pool *pgxpool.Pool
}

// NewOliveBranchRepository creates a new OliveBranchRepository
func NewOliveBranchRepository(pool *pgxpool.Pool) *OliveBranchRepository {
	return &OliveBranchRepository{pool: pool}
}

// OliveBranchListParams contains parameters for listing olive branches
type OliveBranchListParams struct {
	ReceiverID int
	Page       int
	Size       int
	Status     *int
}

// ListByReceiverID retrieves paginated olive branches received by a user
func (r *OliveBranchRepository) ListByReceiverID(ctx context.Context, params OliveBranchListParams) ([]models.OliveBranch, int64, error) {
	// Count total
	countArgs := []interface{}{params.ReceiverID}
	countQuery := `SELECT COUNT(*) FROM olive_branch_record WHERE receiver_id = $1`
	if params.Status != nil {
		countQuery += ` AND status = $2`
		countArgs = append(countArgs, *params.Status)
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count olive branches: %w", err)
	}

	// Query with pagination
	offset := (params.Page - 1) * params.Size
	args := []interface{}{params.ReceiverID}
	argIndex := 2

	query := `
		SELECT 
			ob.id, ob.sender_id, ob.receiver_id, ob.related_project_id,
			ob.type, ob.cost_type, ob.has_sms_notify, ob.message, ob.status,
			ob.created_at, ob.updated_at,
			p.name AS project_name,
			s.id, s.nickname, s.phone, s.email, s.auth_status
		FROM olive_branch_record ob
		LEFT JOIN project p ON ob.related_project_id = p.id
		LEFT JOIN "user" s ON ob.sender_id = s.id
		WHERE ob.receiver_id = $1
	`
	if params.Status != nil {
		query += fmt.Sprintf(` AND ob.status = $%d`, argIndex)
		args = append(args, *params.Status)
		argIndex++
	}
	query += fmt.Sprintf(` ORDER BY ob.created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, params.Size, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query olive branches: %w", err)
	}
	defer rows.Close()

	var records []models.OliveBranch
	for rows.Next() {
		var ob models.OliveBranch
		var sender models.User

		err := rows.Scan(
			&ob.ID, &ob.SenderID, &ob.ReceiverID, &ob.RelatedProjectID,
			&ob.Type, &ob.CostType, &ob.HasSmsNotify, &ob.Message, &ob.Status,
			&ob.CreatedAt, &ob.UpdatedAt,
			&ob.ProjectName,
			&sender.ID, &sender.Nickname, &sender.Phone, &sender.Email, &sender.AuthStatus,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan olive branch: %w", err)
		}
		ob.Sender = &sender
		records = append(records, ob)
	}

	return records, total, nil
}

// GetByID retrieves an olive branch by ID
func (r *OliveBranchRepository) GetByID(ctx context.Context, id int) (*models.OliveBranch, error) {
	query := `
		SELECT 
			ob.id, ob.sender_id, ob.receiver_id, ob.related_project_id,
			ob.type, ob.cost_type, ob.has_sms_notify, ob.message, ob.status,
			ob.created_at, ob.updated_at,
			p.name AS project_name
		FROM olive_branch_record ob
		LEFT JOIN project p ON ob.related_project_id = p.id
		WHERE ob.id = $1
	`

	var ob models.OliveBranch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ob.ID, &ob.SenderID, &ob.ReceiverID, &ob.RelatedProjectID,
		&ob.Type, &ob.CostType, &ob.HasSmsNotify, &ob.Message, &ob.Status,
		&ob.CreatedAt, &ob.UpdatedAt,
		&ob.ProjectName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query olive branch by id: %w", err)
	}

	return &ob, nil
}

// Create creates a new olive branch record
func (r *OliveBranchRepository) Create(ctx context.Context, ob *models.OliveBranch) error {
	query := `
		INSERT INTO olive_branch_record (
			sender_id, receiver_id, related_project_id,
			type, cost_type, has_sms_notify, message, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		ob.SenderID, ob.ReceiverID, ob.RelatedProjectID,
		ob.Type, ob.CostType, ob.HasSmsNotify, ob.Message, ob.Status,
	).Scan(&ob.ID, &ob.CreatedAt, &ob.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create olive branch: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of an olive branch
func (r *OliveBranchRepository) UpdateStatus(ctx context.Context, id int, status int) error {
	query := `UPDATE olive_branch_record SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update olive branch status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("olive branch not found")
	}

	return nil
}
