package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// ListSchools handles GET /dictionaries/schools
func (s *Server) ListSchools(ctx echo.Context, params api.ListSchoolsParams) error {
	schools, err := s.repo.School.List(ctx.Request().Context(), params.Keyword)
	if err != nil {
		return InternalError(ctx, "获取学校列表失败")
	}

	// Convert to VOs
	var schoolVOs []api.SchoolVO
	for _, school := range schools {
		schoolVOs = append(schoolVOs, *school.ToVO())
	}

	return Success(ctx, schoolVOs)
}

// ListMajors handles GET /dictionaries/majors
func (s *Server) ListMajors(ctx echo.Context, params api.ListMajorsParams) error {
	classes, err := s.repo.Major.ListWithMajors(ctx.Request().Context(), params)
	if err != nil {
		return InternalError(ctx, "获取专业列表失败")
	}

	// Convert to VOs
	var classVOs []api.MajorClassVO
	for _, mc := range classes {
		classVOs = append(classVOs, *mc.ToVO())
	}

	return Success(ctx, classVOs)
}
