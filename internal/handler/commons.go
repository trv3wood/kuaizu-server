package handler

import (
	"github.com/labstack/echo/v4"
)

// UploadFile handles POST /commons/uploads
func (s *Server) UploadFile(ctx echo.Context) error {
	file, header, err := ctx.Request().FormFile("file")
	if err != nil {
		return BadRequest(ctx, "缺少文件字段 'file'")
	}
	defer file.Close()

	result, err := s.svc.Commons.UploadFile(file, header)
	if err != nil {
		return mapServiceError(ctx, err)
	}

	return Success(ctx, map[string]string{"url": result.URL})
}
