package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// ApplicationRepository handles project application database operations
type ApplicationRepository struct {
	pool *pgxpool.Pool
}

// NewApplicationRepository creates a new ApplicationRepository
func NewApplicationRepository(pool *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{pool: pool}
}

// ApplicationListParams contains parameters for listing applications
type ApplicationListParams struct {
	Page      int
	Size      int
	ProjectID int
	Status    *int
}

// List retrieves paginated applications for a project with applicant info
func (r *ApplicationRepository) List(ctx context.Context, params ApplicationListParams) ([]models.ProjectApplication, int64, error) {
	// Build WHERE clause
	conditions := []string{"pa.project_id = $1"}
	args := []interface{}{params.ProjectID}
	argIndex := 2

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("pa.status = $%d", argIndex))
		args = append(args, *params.Status)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM project_application pa WHERE %s`, whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	// Query with pagination
	offset := (params.Page - 1) * params.Size
	query := fmt.Sprintf(`
		SELECT
			pa.id, pa.project_id, pa.user_id, pa.apply_reason, pa.contact,
			pa.status, pa.reply_msg, pa.applied_at, pa.updated_at,
			p.name as project_name,
			u.id, u.openid, u.nickname, u.phone, u.email,
			u.school_id, u.major_id, u.grade, u.olive_branch_count,
			u.free_branch_used_today, u.last_active_date,
			u.auth_status, u.auth_img_url, u.created_at
		FROM project_application pa
		LEFT JOIN project p ON pa.project_id = p.id
		LEFT JOIN "user" u ON pa.user_id = u.id
		WHERE %s
		ORDER BY pa.applied_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Size, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query applications: %w", err)
	}
	defer rows.Close()

	var applications []models.ProjectApplication
	for rows.Next() {
		var app models.ProjectApplication
		var applicant models.User
		err := rows.Scan(
			&app.ID, &app.ProjectID, &app.UserID, &app.ApplyReason, &app.Contact,
			&app.Status, &app.ReplyMsg, &app.AppliedAt, &app.UpdatedAt,
			&app.ProjectName,
			&applicant.ID, &applicant.OpenID, &applicant.Nickname, &applicant.Phone, &applicant.Email,
			&applicant.SchoolID, &applicant.MajorID, &applicant.Grade, &applicant.OliveBranchCount,
			&applicant.FreeBranchUsedToday, &applicant.LastActiveDate,
			&applicant.AuthStatus, &applicant.AuthImgUrl, &applicant.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan application: %w", err)
		}
		app.Applicant = &applicant
		applications = append(applications, app)
	}

	return applications, total, nil
}

// Create creates a new application
func (r *ApplicationRepository) Create(ctx context.Context, app *models.ProjectApplication) error {
	query := `
		INSERT INTO project_application (
			project_id, user_id, apply_reason, contact, status
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, applied_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		app.ProjectID, app.UserID, app.ApplyReason, app.Contact, app.Status,
	).Scan(&app.ID, &app.AppliedAt, &app.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	return nil
}

// GetByID retrieves an application by ID
func (r *ApplicationRepository) GetByID(ctx context.Context, id int) (*models.ProjectApplication, error) {
	query := `
		SELECT
			pa.id, pa.project_id, pa.user_id, pa.apply_reason, pa.contact,
			pa.status, pa.reply_msg, pa.applied_at, pa.updated_at
		FROM project_application pa
		WHERE pa.id = $1
	`

	var app models.ProjectApplication
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.ProjectID, &app.UserID, &app.ApplyReason, &app.Contact,
		&app.Status, &app.ReplyMsg, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query application by id: %w", err)
	}

	return &app, nil
}

// CheckDuplicate checks if a user has already applied to a project
func (r *ApplicationRepository) CheckDuplicate(ctx context.Context, projectID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM project_application WHERE project_id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check duplicate application: %w", err)
	}
	return exists, nil
}
