package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/trv3wood/kuaizu-server/api"
)

// ========== Commons Module (Not Implemented) ==========

// UploadFile handles POST /commons/uploads
func (s *Server) UploadFile(ctx echo.Context) error {
	return NotImplemented(ctx)
}

// ========== Orders Module (Not Implemented) ==========

// CreateOrder handles POST /orders
func (s *Server) CreateOrder(ctx echo.Context) error {
	return NotImplemented(ctx)
}

// GetOrder handles GET /orders/{id}
func (s *Server) GetOrder(ctx echo.Context, id int) error {
	return NotImplemented(ctx)
}

// InitiatePayment handles POST /orders/{id}/pay
func (s *Server) InitiatePayment(ctx echo.Context, id int) error {
	return NotImplemented(ctx)
}

// ========== Project Applications Module (Not Implemented) ==========

// ReviewApplication handles PATCH /project-applications/{id}
func (s *Server) ReviewApplication(ctx echo.Context, id int) error {
	return NotImplemented(ctx)
}

// ========== Talent Profiles Module (Not Implemented) ==========

// ListTalentProfiles handles GET /talent-profiles
func (s *Server) ListTalentProfiles(ctx echo.Context, params api.ListTalentProfilesParams) error {
	return NotImplemented(ctx)
}

// UpsertTalentProfile handles POST /talent-profiles
func (s *Server) UpsertTalentProfile(ctx echo.Context) error {
	return NotImplemented(ctx)
}

// GetTalentProfile handles GET /talent-profiles/{id}
func (s *Server) GetTalentProfile(ctx echo.Context, id int) error {
	return NotImplemented(ctx)
}
