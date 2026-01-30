package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// ApplicationRepository handles project application database operations
type ApplicationRepository struct {
	db *sqlx.DB
}

// NewApplicationRepository creates a new ApplicationRepository
func NewApplicationRepository(db *sqlx.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
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
	conditions := []string{"pa.project_id = ?"}
	args := []interface{}{params.ProjectID}

	if params.Status != nil {
		conditions = append(conditions, "pa.status = ?")
		args = append(args, *params.Status)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM project_application pa WHERE %s`, whereClause)
	var total int64
	err := r.db.QueryRowxContext(ctx, countQuery, args...).Scan(&total)
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
		LEFT JOIN `+"`user`"+` u ON pa.user_id = u.id
		WHERE %s
		ORDER BY pa.applied_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, params.Size, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
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
		) VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		app.ProjectID, app.UserID, app.ApplyReason, app.Contact, app.Status,
	)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	app.ID = int(id)

	return nil
}

// GetByID retrieves an application by ID
func (r *ApplicationRepository) GetByID(ctx context.Context, id int) (*models.ProjectApplication, error) {
	query := `
		SELECT
			pa.id, pa.project_id, pa.user_id, pa.apply_reason, pa.contact,
			pa.status, pa.reply_msg, pa.applied_at, pa.updated_at
		FROM project_application pa
		WHERE pa.id = ?
	`

	var app models.ProjectApplication
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&app.ID, &app.ProjectID, &app.UserID, &app.ApplyReason, &app.Contact,
		&app.Status, &app.ReplyMsg, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query application by id: %w", err)
	}

	return &app, nil
}

// CheckDuplicate checks if a user has already applied to a project
func (r *ApplicationRepository) CheckDuplicate(ctx context.Context, projectID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM project_application WHERE project_id = ? AND user_id = ?)`
	var exists bool
	err := r.db.QueryRowxContext(ctx, query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check duplicate application: %w", err)
	}
	return exists, nil
}

// UpdateStatus updates the status and reply message of an application
func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id int, status int, replyMsg *string) error {
	query := `UPDATE project_application SET status = ?, reply_msg = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, status, replyMsg, id)
	if err != nil {
		return fmt.Errorf("update application status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}
