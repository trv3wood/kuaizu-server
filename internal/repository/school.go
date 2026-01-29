package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// SchoolRepository handles school database operations
type SchoolRepository struct {
	pool *pgxpool.Pool
}

// NewSchoolRepository creates a new SchoolRepository
func NewSchoolRepository(pool *pgxpool.Pool) *SchoolRepository {
	return &SchoolRepository{pool: pool}
}

// List retrieves schools with optional keyword search
func (r *SchoolRepository) List(ctx context.Context, keyword *string) ([]*models.School, error) {
	query := `
		SELECT id, school_name, school_code, created_at, updated_at
		FROM school
	`
	args := []interface{}{}

	if keyword != nil && *keyword != "" {
		query += ` WHERE school_name ILIKE $1 OR school_code ILIKE $1`
		searchPattern := "%" + *keyword + "%"
		args = append(args, searchPattern)
	}

	query += ` ORDER BY school_name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schools: %w", err)
	}
	defer rows.Close()

	var schools []*models.School
	for rows.Next() {
		var school models.School
		err := rows.Scan(
			&school.ID,
			&school.SchoolName,
			&school.SchoolCode,
			&school.CreatedAt,
			&school.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan school: %w", err)
		}
		schools = append(schools, &school)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schools: %w", err)
	}

	return schools, nil
}
