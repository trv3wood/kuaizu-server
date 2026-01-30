package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by ID with joined school and major info
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT
			u.id, u.openid, u.nickname, u.phone, u.email,
			u.school_id, u.major_id, u.grade, u.olive_branch_count,
			u.free_branch_used_today, u.last_active_date,
			u.auth_status, u.auth_img_url, u.created_at,
			s.school_name, s.school_code,
			m.major_name, m.class_id
		FROM ` + "`user`" + ` u
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE u.id = ?
	`

	var user models.User
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&user.ID, &user.OpenID, &user.Nickname, &user.Phone, &user.Email,
		&user.SchoolID, &user.MajorID, &user.Grade, &user.OliveBranchCount,
		&user.FreeBranchUsedToday, &user.LastActiveDate,
		&user.AuthStatus, &user.AuthImgUrl, &user.CreatedAt,
		&user.SchoolName, &user.SchoolCode,
		&user.MajorName, &user.ClassID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &user, nil
}

// GetByOpenID retrieves a user by WeChat OpenID
func (r *UserRepository) GetByOpenID(ctx context.Context, openid string) (*models.User, error) {
	query := `
		SELECT
			u.id, u.openid, u.nickname, u.phone, u.email,
			u.school_id, u.major_id, u.grade, u.olive_branch_count,
			u.free_branch_used_today, u.last_active_date,
			u.auth_status, u.auth_img_url, u.created_at,
			s.school_name, s.school_code,
			m.major_name, m.class_id
		FROM ` + "`user`" + ` u
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE u.openid = ?
	`

	var user models.User
	err := r.db.QueryRowxContext(ctx, query, openid).Scan(
		&user.ID, &user.OpenID, &user.Nickname, &user.Phone, &user.Email,
		&user.SchoolID, &user.MajorID, &user.Grade, &user.OliveBranchCount,
		&user.FreeBranchUsedToday, &user.LastActiveDate,
		&user.AuthStatus, &user.AuthImgUrl, &user.CreatedAt,
		&user.SchoolName, &user.SchoolCode,
		&user.MajorName, &user.ClassID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by openid: %w", err)
	}

	return &user, nil
}

// Create creates a new user and returns the created user
func (r *UserRepository) Create(ctx context.Context, openid string) (*models.User, error) {
	query := `
		INSERT INTO ` + "`user`" + ` (openid, olive_branch_count, free_branch_used_today, auth_status)
		VALUES (?, 0, 0, 0)
	`

	result, err := r.db.ExecContext(ctx, query, openid)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return r.GetByID(ctx, int(id))
}

// Update updates user fields
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE ` + "`user`" + ` SET
			nickname = ?,
			phone = ?,
			email = ?,
			school_id = ?,
			major_id = ?,
			grade = ?,
			auth_img_url = ?,
			auth_status = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Nickname,
		user.Phone,
		user.Email,
		user.SchoolID,
		user.MajorID,
		user.Grade,
		user.AuthImgUrl,
		user.AuthStatus,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// UpdateQuota updates user's olive branch quota fields
func (r *UserRepository) UpdateQuota(ctx context.Context, user *models.User) error {
	query := `
		UPDATE ` + "`user`" + ` SET
			olive_branch_count = ?,
			free_branch_used_today = ?,
			last_active_date = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		user.OliveBranchCount,
		user.FreeBranchUsedToday,
		user.LastActiveDate,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user quota: %w", err)
	}

	return nil
}
