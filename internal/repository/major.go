package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/models"
)

// MajorRepository handles major database operations
type MajorRepository struct {
	db *sqlx.DB
}

// NewMajorRepository creates a new MajorRepository
func NewMajorRepository(db *sqlx.DB) *MajorRepository {
	return &MajorRepository{db: db}
}

func (r *MajorRepository) List(ctx context.Context, params *api.ListMajorsParams) ([]models.Major, error) {
	query := `
		SELECT id, major_name, class_id
		FROM major
		WHERE class_id = ?
	`
	rows, err := r.db.QueryxContext(ctx, query, params.ClassId)
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

	if params.ClassId != nil {
		conditions = append(conditions, "id = ?")
		args = append(args, *params.ClassId)
	}

	if params.ClassKeyword != nil && *params.ClassKeyword != "" {
		conditions = append(conditions, "class_name LIKE ?")
		args = append(args, "%"+*params.ClassKeyword+"%")
	}

	classQuery := "SELECT id, class_name FROM major_class"
	if len(conditions) > 0 {
		classQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	classQuery += " ORDER BY id"

	rows, err := r.db.QueryxContext(ctx, classQuery, args...)
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

	// Use sqlx.In to expand the slice to placeholders
	majorQuery := `SELECT id, major_name, class_id FROM major WHERE class_id IN (?)`
	majorArgs := []interface{}{classIDs}

	if params.MajorKeyword != nil && *params.MajorKeyword != "" {
		majorQuery = `SELECT id, major_name, class_id FROM major WHERE class_id IN (?) AND major_name LIKE ?`
		majorArgs = append(majorArgs, "%"+*params.MajorKeyword+"%")
	}

	// Expand IN clause
	majorQuery, majorArgs, err = sqlx.In(majorQuery, majorArgs...)
	if err != nil {
		return nil, fmt.Errorf("expand IN clause: %w", err)
	}
	majorQuery = r.db.Rebind(majorQuery)
	majorQuery += " ORDER BY class_id, id"

	mRows, err := r.db.QueryxContext(ctx, majorQuery, majorArgs...)
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
