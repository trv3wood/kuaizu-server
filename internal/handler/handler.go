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
func GetUserID(ctx interface{ Get(string) interface{} }) int {
	userID, ok := ctx.Get("userID").(int)
	if !ok {
		// This should never happen if JWT middleware is properly configured
		panic("userID not found in context - auth middleware not applied?")
	}
	return userID
}

// GetOpenID extracts OpenID from context (set by auth middleware)
func GetOpenID(ctx interface{ Get(string) interface{} }) string {
	openID, ok := ctx.Get("openID").(string)
	if !ok {
		return ""
	}
	return openID
}
