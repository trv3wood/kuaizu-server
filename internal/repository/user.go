package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
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
		FROM "user" u
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE u.id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.OpenID, &user.Nickname, &user.Phone, &user.Email,
		&user.SchoolID, &user.MajorID, &user.Grade, &user.OliveBranchCount,
		&user.FreeBranchUsedToday, &user.LastActiveDate,
		&user.AuthStatus, &user.AuthImgUrl, &user.CreatedAt,
		&user.SchoolName, &user.SchoolCode,
		&user.MajorName, &user.ClassID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
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
		FROM "user" u
		LEFT JOIN school s ON u.school_id = s.id
		LEFT JOIN major m ON u.major_id = m.id
		WHERE u.openid = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, openid).Scan(
		&user.ID, &user.OpenID, &user.Nickname, &user.Phone, &user.Email,
		&user.SchoolID, &user.MajorID, &user.Grade, &user.OliveBranchCount,
		&user.FreeBranchUsedToday, &user.LastActiveDate,
		&user.AuthStatus, &user.AuthImgUrl, &user.CreatedAt,
		&user.SchoolName, &user.SchoolCode,
		&user.MajorName, &user.ClassID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by openid: %w", err)
	}

	return &user, nil
}

// Create creates a new user and returns the created user
func (r *UserRepository) Create(ctx context.Context, openid string) (*models.User, error) {
	query := `
		INSERT INTO "user" (openid, olive_branch_count, free_branch_used_today, auth_status)
		VALUES ($1, 0, 0, 0)
		RETURNING id, openid, nickname, phone, email, school_id, major_id, grade,
			olive_branch_count, free_branch_used_today, last_active_date,
			auth_status, auth_img_url, created_at
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, openid).Scan(
		&user.ID, &user.OpenID, &user.Nickname, &user.Phone, &user.Email,
		&user.SchoolID, &user.MajorID, &user.Grade, &user.OliveBranchCount,
		&user.FreeBranchUsedToday, &user.LastActiveDate,
		&user.AuthStatus, &user.AuthImgUrl, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

// Update updates user fields
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE "user" SET
			nickname = $2,
			phone = $3,
			email = $4,
			school_id = $5,
			major_id = $6,
			grade = $7,
			auth_img_url = $8,
			auth_status = $9
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Nickname,
		user.Phone,
		user.Email,
		user.SchoolID,
		user.MajorID,
		user.Grade,
		user.AuthImgUrl,
		user.AuthStatus,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}
