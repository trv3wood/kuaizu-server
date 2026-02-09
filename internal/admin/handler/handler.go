package handler

import (
	"github.com/trv3wood/kuaizu-server/internal/repository"
)

// AdminServer handles admin API requests
type AdminServer struct {
	repo *repository.Repository
}

// NewAdminServer creates a new AdminServer instance
func NewAdminServer(repo *repository.Repository) *AdminServer {
	return &AdminServer{repo: repo}
}
