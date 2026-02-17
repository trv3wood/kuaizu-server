package handler

import (
	"github.com/labstack/echo/v4"
)

// ========== Commons Module (Not Implemented) ==========

// UploadFile handles POST /commons/uploads
func (s *Server) UploadFile(ctx echo.Context) error {
	return NotImplemented(ctx)
}
