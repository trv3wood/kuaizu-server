package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// TalentProfileRepository handles talent profile database operations
type TalentProfileRepository struct {
	db *sqlx.DB
}

// NewTalentProfileRepository creates a new TalentProfileRepository
func NewTalentProfileRepository(db *sqlx.DB) *TalentProfileRepository {
	return &TalentProfileRepository{db: db}
}

// TalentProfileListParams contains parameters for listing talent profiles
type TalentProfileListParams struct {
	Page     int
	Size     int
	SchoolID *int
	MajorID  *int
	Keyword  *string
	Status   *int
}

// List retrieves paginated talent profiles with optional filters
func (r *TalentProfileRepository) List(ctx context.Context, params TalentProfileListParams) ([]models.TalentProfile, int64, error) {
	// Build WHERE clause - only show active profiles
	conditions := []string{"tp.status = 1"}
	args := []interface{}{}

	if params.SchoolID != nil {
		conditions = append(conditions, "u.school_id = ?")
		args = append(args, *params.SchoolID)
	}

	if params.MajorID != nil {
		conditions = append(conditions, "u.major_id = ?")
		args = append(args, *params.MajorID)
	}

	if params.Keyword != nil && *params.Keyword != "" {
		conditions = append(conditions, "(u.nickname LIKE ? OR tp.self_evaluation LIKE ? OR tp.skill_summary LIKE ?)")
		pattern := "%" + *params.Keyword + "%"
		args = append(args, pattern, pattern, pattern)
	}

	if params.Status != nil {
		conditions = append(conditions, "tp.status = ?")
		args = append(args, *params.Status)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM talent_profile tp
		LEFT JOIN `+"`user`"+` u ON tp.user_id = u.id
		WHERE %s
	`, whereClause)
	var total int64
	if err := r.db.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count talent profiles: %w", err)
	}

	// Query with pagination
	offset := (params.Page - 1) * params.Size
	query := fmt.Sprintf(`
		SELECT 
			tp.id, tp.user_id, tp.self_evaluation, tp.skill_summary,
			tp.project_experience, tp.mbti, tp.status, tp.is_public_contact,
			tp.created_at, tp.updated_at,
			u.nickname, s.school_name, m.major_name, u.phone, u.email, u.avatar_url
		FROM talent_profile tp
		LEFT JOIN `+"`user`"+` u ON tp.user_id = u.id
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE %s
		ORDER BY tp.updated_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, params.Size, offset)

	var profiles []models.TalentProfile
	if err := r.db.SelectContext(ctx, &profiles, query, args...); err != nil {
		return nil, 0, fmt.Errorf("query talent profiles: %w", err)
	}

	return profiles, total, nil
}

// GetByID retrieves a talent profile by ID with user info
func (r *TalentProfileRepository) GetByID(ctx context.Context, id int) (*models.TalentProfile, error) {
	query := `
		SELECT 
			tp.id, tp.user_id, tp.self_evaluation, tp.skill_summary,
			tp.project_experience, tp.mbti, tp.status, tp.is_public_contact,
			tp.created_at, tp.updated_at,
			u.nickname, s.school_name, m.major_name, u.phone, u.email
		FROM talent_profile tp
		LEFT JOIN ` + "`user`" + ` u ON tp.user_id = u.id
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE tp.id = ?
	`

	var p models.TalentProfile
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&p); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query talent profile by id: %w", err)
	}

	return &p, nil
}

// GetByUserID retrieves a talent profile by user ID
func (r *TalentProfileRepository) GetByUserID(ctx context.Context, userID int) (*models.TalentProfile, error) {
	query := `
		SELECT 
			tp.id, tp.user_id, tp.self_evaluation, tp.skill_summary,
			tp.project_experience, tp.mbti, tp.status, tp.is_public_contact,
			tp.created_at, tp.updated_at,
			u.nickname, s.school_name, m.major_name, u.phone, u.email
		FROM talent_profile tp
		LEFT JOIN ` + "`user`" + ` u ON tp.user_id = u.id
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE tp.user_id = ?
	`

	var p models.TalentProfile
	if err := r.db.QueryRowxContext(ctx, query, userID).StructScan(&p); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query talent profile by user id: %w", err)
	}

	return &p, nil
}

// Upsert creates or updates a talent profile for a user
func (r *TalentProfileRepository) Upsert(ctx context.Context, p *models.TalentProfile) error {
	// Check if profile exists
	existing, err := r.GetByUserID(ctx, p.UserID)
	if err != nil {
		return err
	}

	if existing == nil {
		// Insert
		query := `
			INSERT INTO talent_profile (
				user_id, self_evaluation, skill_summary, project_experience,
				mbti, status, is_public_contact
			) VALUES (
				:user_id, :self_evaluation, :skill_summary, :project_experience,
				:mbti, :status, :is_public_contact
			)
		`
		result, err := r.db.NamedExecContext(ctx, query, p)
		if err != nil {
			return fmt.Errorf("insert talent profile: %w", err)
		}
		id, _ := result.LastInsertId()
		p.ID = int(id)
	} else {
		// Update
		query := `
			UPDATE talent_profile SET
				self_evaluation = :self_evaluation,
				skill_summary = :skill_summary,
				project_experience = :project_experience,
				mbti = :mbti,
				status = :status,
				is_public_contact = :is_public_contact,
				updated_at = CURRENT_TIMESTAMP
			WHERE user_id = :user_id
		`
		_, err := r.db.NamedExecContext(ctx, query, p)
		if err != nil {
			return fmt.Errorf("update talent profile: %w", err)
		}
		p.ID = existing.ID
	}

	return nil
}

// DeleteByUserID deletes a talent profile by user ID
func (r *TalentProfileRepository) DeleteByUserID(ctx context.Context, userID int) error {
	query := `
		UPDATE talent_profile SET status = 0, is_public_contact = 0 WHERE user_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete talent profile by user id: %w", err)
	}
	return nil
}
