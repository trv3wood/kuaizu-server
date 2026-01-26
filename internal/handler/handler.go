package handler

import (
	"github.com/trv3wood/kuaizu-server/api"
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// Server implements api.ServerInterface
type Server struct {
	repo *repository.Repository
}

// Ensure Server implements the generated ServerInterface
var _ api.ServerInterface = (*Server)(nil)

// NewServer creates a new Server instance
func NewServer(repo *repository.Repository) *Server {
	return &Server{repo: repo}
}

// GetUserID extracts user ID from context (set by auth middleware)
// For prototype, returns a mock user ID if not authenticated
func GetUserID(ctx interface{ Get(string) interface{} }) int {
	if userID, ok := ctx.Get("userID").(int); ok {
		return userID
	}
	// Prototype: return mock user ID 1 for testing
	return 1
}
