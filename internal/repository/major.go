package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// MajorRepository handles major database operations
type MajorRepository struct {
	pool *pgxpool.Pool
}

// NewMajorRepository creates a new MajorRepository
func NewMajorRepository(pool *pgxpool.Pool) *MajorRepository {
	return &MajorRepository{pool: pool}
}

func (r *MajorRepository) List(ctx context.Context, params *api.ListMajorsParams) ([]models.Major, error) {
	query := `
		SELECT id, major_name, class_id
		FROM major
		WHERE class_id = $1
	`
	rows, err := r.pool.Query(ctx, query, params.ClassId)
	if err != nil {
		return nil, fmt.Errorf("query majors: %w", err)
	}
	defer rows.Close()
	var majors []models.Major
	for rows.Next() {
		var major models.Major
		err := rows.Scan(&major.Id, &major.MajorName, &major.ClassId)
		if err != nil {
			return nil, fmt.Errorf("scan majors: %w", err)
		}
		majors = append(majors, major)
	}
	return majors, nil
}

// ListWithMajors returns major classes with their majors
func (r *MajorRepository) ListWithMajors(ctx context.Context, params api.ListMajorsParams) ([]models.MajorClass, error) {
	// 1. Query classes
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.ClassId != nil {
		conditions = append(conditions, fmt.Sprintf("id = $%d", argIndex))
		args = append(args, *params.ClassId)
		argIndex++
	}

	if params.ClassKeyword != nil && *params.ClassKeyword != "" {
		conditions = append(conditions, fmt.Sprintf("class_name ILIKE $%d", argIndex))
		args = append(args, "%"+*params.ClassKeyword+"%")
		argIndex++
	}

	classQuery := "SELECT id, class_name FROM major_class"
	if len(conditions) > 0 {
		classQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	classQuery += " ORDER BY id"

	rows, err := r.pool.Query(ctx, classQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query major classes: %w", err)
	}
	defer rows.Close()

	var classes []models.MajorClass
	classIDMap := make(map[int]*models.MajorClass)

	for rows.Next() {
		var mc models.MajorClass
		if err := rows.Scan(&mc.Id, &mc.ClassName); err != nil {
			return nil, fmt.Errorf("scan major class: %w", err)
		}
		classes = append(classes, mc)
	}

	if len(classes) == 0 {
		return classes, nil
	}

	// 2. Query majors for these classes
	classIDs := make([]int, len(classes))
	for i, c := range classes {
		classIDs[i] = c.Id
	}

	majorConditions := []string{"class_id = ANY($1)"}
	majorArgs := []interface{}{classIDs}
	mArgIndex := 2

	if params.MajorKeyword != nil && *params.MajorKeyword != "" {
		majorConditions = append(majorConditions, fmt.Sprintf("major_name ILIKE $%d", mArgIndex))
		majorArgs = append(majorArgs, "%"+*params.MajorKeyword+"%")
		mArgIndex++
	}

	majorQuery := fmt.Sprintf(`
		SELECT id, major_name, class_id 
		FROM major 
		WHERE %s
		ORDER BY class_id, id
	`, strings.Join(majorConditions, " AND "))

	mRows, err := r.pool.Query(ctx, majorQuery, majorArgs...)
	if err != nil {
		return nil, fmt.Errorf("query majors: %w", err)
	}
	defer mRows.Close()

	// Use map to easily append majors to classes
	for i := range classes {
		classIDMap[classes[i].Id] = &classes[i]
	}

	for mRows.Next() {
		var m models.Major
		if err := mRows.Scan(&m.Id, &m.MajorName, &m.ClassId); err != nil {
			return nil, fmt.Errorf("scan major: %w", err)
		}
		if mc, ok := classIDMap[m.ClassId]; ok {
			mc.Majors = append(mc.Majors, m)
		}
	}

	// 3. Filter out classes that have no majors (if filtering by major keyword)
	if params.MajorKeyword != nil && *params.MajorKeyword != "" {
		var filteredClasses []models.MajorClass
		for _, c := range classes {
			if len(c.Majors) > 0 {
				filteredClasses = append(filteredClasses, c)
			}
		}
		return filteredClasses, nil
	}

	return classes, nil
}
