package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// SchoolRepository handles school database operations
type SchoolRepository struct {
	db *sqlx.DB
}

// NewSchoolRepository creates a new SchoolRepository
func NewSchoolRepository(db *sqlx.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

// List retrieves schools with optional keyword search
func (r *SchoolRepository) List(ctx context.Context, keyword *string) ([]*models.School, error) {
	query := `
		SELECT id, school_name, school_code, created_at, updated_at
		FROM school
	`
	args := []interface{}{}

	if keyword != nil && *keyword != "" {
		query += ` WHERE school_name LIKE ? OR school_code LIKE ?`
		searchPattern := "%" + *keyword + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += ` ORDER BY school_name ASC`

	var schools []*models.School
	err := r.db.SelectContext(ctx, &schools, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schools: %w", err)
	}

	return schools, nil
}
