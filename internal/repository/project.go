package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// ProjectRepository handles project database operations
type ProjectRepository struct {
	pool *pgxpool.Pool
}

// NewProjectRepository creates a new ProjectRepository
func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

// ListParams contains parameters for listing projects
type ListParams struct {
	Page      int
	Size      int
	Keyword   *string
	SchoolID  *int
	Status    *int
	Direction *int
}

// List retrieves paginated projects with optional filters
func (r *ProjectRepository) List(ctx context.Context, params ListParams) ([]models.Project, int64, error) {
	// Build WHERE clause
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if params.Keyword != nil && *params.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*params.Keyword+"%")
		argIndex++
	}

	if params.SchoolID != nil {
		conditions = append(conditions, fmt.Sprintf("p.school_id = $%d", argIndex))
		args = append(args, *params.SchoolID)
		argIndex++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", argIndex))
		args = append(args, *params.Status)
		argIndex++
	}

	if params.Direction != nil {
		conditions = append(conditions, fmt.Sprintf("p.direction = $%d", argIndex))
		args = append(args, *params.Direction)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM project p WHERE %s`, whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	// Query with pagination
	offset := (params.Page - 1) * params.Size
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.creator_id, p.name, p.description, p.school_id,
			p.direction, p.member_count, p.status,
			p.promotion_status, p.promotion_expire_time, p.view_count,
			p.created_at, p.updated_at,
			s.school_name
		FROM project p
		LEFT JOIN school s ON p.school_id = s.id
		WHERE %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Size, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(
			&p.ID, &p.CreatorID, &p.Name, &p.Description, &p.SchoolID,
			&p.Direction, &p.MemberCount, &p.Status,
			&p.PromotionStatus, &p.PromotionExpireTime, &p.ViewCount,
			&p.CreatedAt, &p.UpdatedAt,
			&p.SchoolName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, total, nil
}

// GetByID retrieves a project by ID with creator info
func (r *ProjectRepository) GetByID(ctx context.Context, id int) (*models.Project, error) {
	query := `
		SELECT 
			p.id, p.creator_id, p.name, p.description, p.school_id,
			p.direction, p.member_count, p.status,
			p.promotion_status, p.promotion_expire_time, p.view_count,
			p.created_at, p.updated_at,
			s.school_name,
			u.id, u.openid, u.nickname, u.phone, u.email,
			u.school_id, u.major_id, u.grade, u.olive_branch_count,
			u.free_branch_used_today, u.last_active_date,
			u.auth_status, u.auth_img_url, u.created_at
		FROM project p
		LEFT JOIN school s ON p.school_id = s.id
		LEFT JOIN "user" u ON p.creator_id = u.id
		WHERE p.id = $1
	`

	var p models.Project
	var creator models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CreatorID, &p.Name, &p.Description, &p.SchoolID,
		&p.Direction, &p.MemberCount, &p.Status,
		&p.PromotionStatus, &p.PromotionExpireTime, &p.ViewCount,
		&p.CreatedAt, &p.UpdatedAt,
		&p.SchoolName,
		&creator.ID, &creator.OpenID, &creator.Nickname, &creator.Phone, &creator.Email,
		&creator.SchoolID, &creator.MajorID, &creator.Grade, &creator.OliveBranchCount,
		&creator.FreeBranchUsedToday, &creator.LastActiveDate,
		&creator.AuthStatus, &creator.AuthImgUrl, &creator.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query project by id: %w", err)
	}

	p.Creator = &creator
	return &p, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, p *models.Project) error {
	query := `
		INSERT INTO project (
			creator_id, name, description, school_id, direction,
			member_count, status, promotion_status, view_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		p.CreatorID, p.Name, p.Description, p.SchoolID, p.Direction,
		p.MemberCount, p.Status, p.PromotionStatus, p.ViewCount,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	return nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, p *models.Project) error {
	query := `
		UPDATE project SET
			name = $2,
			description = $3,
			direction = $4,
			member_count = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.Description, p.Direction, p.MemberCount,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// Delete performs a logical delete (sets status to CLOSED)
func (r *ProjectRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE project SET status = 3, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// IsOwner checks if a user is the creator of a project
func (r *ProjectRepository) IsOwner(ctx context.Context, projectID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM project WHERE id = $1 AND creator_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check project owner: %w", err)
	}
	return exists, nil
}

// IncrementViewCount increments the view count of a project
func (r *ProjectRepository) IncrementViewCount(ctx context.Context, id int) error {
	query := `UPDATE project SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment view count: %w", err)
	}
	return nil
}
